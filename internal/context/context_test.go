package context_test

import (
	"context"
	tlsX509 "crypto/x509"
	tlsErrors "errors"
	tlsIO "io"
	tlsLog "log"
	"net/http"
	"net/http/httptest"
	"sync"
	tlsAtomic "sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	tlsYAML "gopkg.in/yaml.v2"

	tlsOpenAPI "github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	tlsNFM "github.com/free5gc/openapi/nrf/NFMgmt"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/sbi/processor"
	tlsFactory "github.com/free5gc/udr/pkg/factory"
)

const (
	concurrentWorkers   = 8
	iterationsPerWorker = 64
)

func newSubscriptionsRouter(p *processor.Processor) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/nudr-dr/v1/subscription-data/subs-to-notify",
		func(c *gin.Context) {
			p.PostSubscriptionDataSubscriptionsProcedure(
				c, models.Udr_DR_SubscriptionDataSubscriptions{})
		})
	return router
}

func postSubscriptionDataSubscriptionsOnce(
	router *gin.Engine, t *testing.T,
) {
	t.Helper()

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost,
		"/nudr-dr/v1/subscription-data/subs-to-notify", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handler status = %d, want 201", w.Code)
	}
}

func TestPostSubscriptionDataSubscriptionsConcurrentAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	self := udr_context.GetSelf()
	self.SubscriptionDataSubscriptions = make(
		map[string]*udr_context.SubscriptionDataSubscriptionRecord)

	p := &processor.Processor{}
	router := newSubscriptionsRouter(p)

	var wg sync.WaitGroup
	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerWorker; j++ {
				postSubscriptionDataSubscriptionsOnce(router, t)
			}
		}()
	}
	wg.Wait()

	total := concurrentWorkers * iterationsPerWorker
	if got := len(self.SubscriptionDataSubscriptions); got != total {
		t.Fatalf("stored subscriptions = %d, want %d: subscription IDs collided "+
			"under concurrency (unsynchronized ID generator / map writes)",
			got, total)
	}
}

// Exercise parsed NF config through the same startup configuration API and
// generated NRF client used in production. No deployment files are modified.
func TestSBIConfigTLSVerification(t *testing.T) {
	t.Cleanup(func() {
		if err := tlsOpenAPI.ConfigureHTTP2Clients(
			tlsOpenAPI.TransportReadIdleTimeout, tlsOpenAPI.TransportPingTimeout, false,
		); err != nil {
			t.Errorf("restore HTTP clients: %v", err)
		}
	})
	var requests tlsAtomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if _, err := tlsIO.WriteString(w, `{}`); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	server.EnableHTTP2 = true
	server.Config.ErrorLog = tlsLog.New(tlsIO.Discard, "", 0)
	server.StartTLS()
	defer server.Close()
	for _, value := range []string{"", "false", "true"} {
		t.Run("verify="+value, func(t *testing.T) {
			data := "configuration:\n  sbi:\n    scheme: https\n"
			if value != "" {
				data += "    tls:\n      verify: " + value + "\n"
			}
			var cfg tlsFactory.Config
			if err := tlsYAML.Unmarshal([]byte(data), &cfg); err != nil {
				t.Fatal(err)
			}
			if cfg.GetTLSVerify() != (value == "true") {
				t.Fatal("verify parsing mismatch")
			}
			if err := tlsOpenAPI.ConfigureHTTP2Clients(
				cfg.GetReadIdleTimeout(), cfg.GetPingTimeout(), cfg.GetTLSVerify(),
			); err != nil {
				t.Fatal(err)
			}
			apiCfg := tlsNFM.NewConfiguration()
			apiCfg.SetBasePath(server.URL)
			client := tlsNFM.NewAPIClient(apiCfg)
			req := &tlsNFM.GetNFInstanceRequest{}
			req.SetNfInstanceID("tls-test")
			before := requests.Load()
			_, err := client.NFInstanceIDDocumentApi.GetNFInstance(context.Background(), req)
			if value == "true" {
				var unknown tlsX509.UnknownAuthorityError
				if !tlsErrors.As(err, &unknown) {
					t.Fatalf("expected unknown authority, got %v", err)
				}
				if requests.Load() != before {
					t.Fatal("HTTP handler reached despite rejected certificate")
				}
			} else if err != nil {
				t.Fatal(err)
			} else if requests.Load() != before+1 {
				t.Fatal("request not received")
			}
		})
	}
}
