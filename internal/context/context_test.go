package context_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/openapi/models"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/sbi/processor"
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
