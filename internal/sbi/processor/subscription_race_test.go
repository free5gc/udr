package processor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/free5gc/openapi/models"
	udr_context "github.com/free5gc/udr/internal/context"
)

const (
	raceWorkers = 48
	raceRounds  = 200
)

func init() {
	gin.SetMode(gin.TestMode)
}

func resetUESubsCollection() {
	udrSelf := udr_context.GetSelf()
	udrSelf.UESubsCollection.Range(func(key, _ interface{}) bool {
		udrSelf.UESubsCollection.Delete(key)
		return true
	})
}

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	return c, w
}

// Concurrent POST/DELETE of SDM subscriptions for the same UE must neither
// crash ("concurrent map writes") nor hand out duplicate subscription IDs.
func TestSdmSubscriptionsConcurrentCreateDelete(t *testing.T) {
	resetUESubsCollection()
	p := &Processor{}
	const ueId = "imsi-208930000000001"

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		ids = make(map[string]struct{})
	)
	for i := 0; i < raceWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < raceRounds; j++ {
				c, w := newTestContext()
				p.CreateSdmSubscriptionsProcedure(c, models.Udr_DR_SdmSubscription{}, "", ueId)
				c.Writer.WriteHeaderNow()
				if w.Code != http.StatusCreated {
					t.Errorf("create: unexpected status %d", w.Code)
					return
				}
				loc := w.Header().Get("Location")
				subsId := loc[lastSlash(loc)+1:]

				mu.Lock()
				if _, dup := ids[subsId]; dup {
					t.Errorf("duplicate subscription ID %s", subsId)
				}
				ids[subsId] = struct{}{}
				mu.Unlock()

				c, _ = newTestContext()
				p.QuerysdmsubscriptionsProcedure(c, ueId)
				c, w = newTestContext()
				p.RemovesdmSubscriptionsProcedure(c, ueId, subsId)
				c.Writer.WriteHeaderNow()
				assert.Equal(t, http.StatusNoContent, w.Code)
			}
		}()
	}
	wg.Wait()
}

// Same for EE subscriptions.
func TestEeSubscriptionsConcurrentCreateDelete(t *testing.T) {
	resetUESubsCollection()
	p := &Processor{}
	const ueId = "imsi-208930000000002"

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		ids = make(map[string]struct{})
	)
	for i := 0; i < raceWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < raceRounds; j++ {
				c, w := newTestContext()
				p.CreateEeSubscriptionsProcedure(c, ueId, models.Udr_DR_EeSubscription{})
				loc := w.Header().Get("Location")
				subsId := loc[lastSlash(loc)+1:]

				mu.Lock()
				if _, dup := ids[subsId]; dup {
					t.Errorf("duplicate subscription ID %s", subsId)
				}
				ids[subsId] = struct{}{}
				mu.Unlock()

				c, _ = newTestContext()
				p.QueryeesubscriptionsProcedure(c, ueId)
				c, _ = newTestContext()
				p.RemoveeeSubscriptionsProcedure(c, ueId, subsId)
			}
		}()
	}
	wg.Wait()
}

func lastSlash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}
