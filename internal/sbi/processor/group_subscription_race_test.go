package processor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/free5gc/openapi/models"
	udr_context "github.com/free5gc/udr/internal/context"
)

// Concurrent first-time creates must share one group and retain every subscription.
func TestEeGroupSubscriptionsConcurrentCreate(t *testing.T) {
	for _, groupID := range []string{"extgroupid-concurrent@test", "anyUE"} {
		t.Run(groupID, func(t *testing.T) {
			udrSelf := udr_context.GetSelf()
			udrSelf.UEGroupCollection.Delete(groupID)
			t.Cleanup(func() { udrSelf.UEGroupCollection.Delete(groupID) })
			p := &Processor{}
			ids := make([]string, raceWorkers)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for i := 0; i < raceWorkers; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					<-start
					c, w := newTestContext()
					p.CreateEeGroupSubscriptionsProcedure(c, groupID, models.Udr_DR_EeSubscription{
						CallbackReference: fmt.Sprintf("https://example.com/callback/%d", i),
					})
					assert.Equal(t, http.StatusCreated, w.Code)
					loc := w.Header().Get("Location")
					ids[i] = loc[lastSlash(loc)+1:]
				}(i)
			}
			close(start)
			wg.Wait()

			seen := make(map[string]bool)
			for _, id := range ids {
				assert.NotEmpty(t, id)
				assert.False(t, seen[id], "duplicate subscription ID %s", id)
				seen[id] = true
			}
			c, w := newTestContext()
			p.QueryEeGroupSubscriptionsProcedure(c, groupID)
			require.Equal(t, http.StatusOK, w.Code)
			var subscriptions []models.Udr_DR_EeSubscription
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &subscriptions))
			require.Len(t, subscriptions, raceWorkers)
			callbacks := make([]string, 0, len(subscriptions))
			for _, subscription := range subscriptions {
				callbacks = append(callbacks, subscription.CallbackReference)
			}
			for i := 0; i < raceWorkers; i++ {
				assert.Contains(t, callbacks, fmt.Sprintf("https://example.com/callback/%d", i))
			}
		})
	}
}

// Each worker owns its subscription while all workers read and write the same group map.
func TestEeGroupSubscriptionsConcurrentCreateUpdateQueryDelete(t *testing.T) {
	const groupID = "extgroupid-mixed@test"
	udrSelf := udr_context.GetSelf()
	udrSelf.UEGroupCollection.Delete(groupID)
	t.Cleanup(func() { udrSelf.UEGroupCollection.Delete(groupID) })
	p := &Processor{}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < raceWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 20; j++ {
				c, w := newTestContext()
				p.CreateEeGroupSubscriptionsProcedure(c, groupID, models.Udr_DR_EeSubscription{})
				assert.Equal(t, http.StatusCreated, w.Code)
				loc := w.Header().Get("Location")
				id := loc[lastSlash(loc)+1:]
				updated := models.Udr_DR_EeSubscription{CallbackReference: "https://example.com/" + id}

				c, w = newTestContext()
				p.UpdateEeGroupSubscriptionsProcedure(c, groupID, id, updated)
				c.Writer.WriteHeaderNow()
				assert.Equal(t, http.StatusNoContent, w.Code)

				c, w = newTestContext()
				p.QueryEeGroupSubscriptionsProcedure(c, groupID)
				assert.Equal(t, http.StatusOK, w.Code)
				var subscriptions []models.Udr_DR_EeSubscription
				if assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &subscriptions)) {
					assert.Contains(t, subscriptions, updated)
				}

				c, w = newTestContext()
				p.RemoveEeGroupSubscriptionsProcedure(c, groupID, id)
				c.Writer.WriteHeaderNow()
				assert.Equal(t, http.StatusNoContent, w.Code)

				c, w = newTestContext()
				p.UpdateEeGroupSubscriptionsProcedure(c, groupID, id, updated)
				assert.Equal(t, http.StatusNotFound, w.Code)
				c, w = newTestContext()
				p.RemoveEeGroupSubscriptionsProcedure(c, groupID, id)
				assert.Equal(t, http.StatusNotFound, w.Code)
			}
		}()
	}
	close(start)
	wg.Wait()

	value, ok := udrSelf.UEGroupCollection.Load(groupID)
	require.True(t, ok)
	group := value.(*udr_context.UEGroupSubsData)
	group.RLock()
	defer group.RUnlock()
	assert.Empty(t, group.EeSubscriptions)
}
