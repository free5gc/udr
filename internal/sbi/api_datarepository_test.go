package sbi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/udr/pkg/factory"
)

// A request that matches the '/application-data/influenceData/:influenceId/:subscriptionId'
// pattern with an influenceId other than "subs-to-notify" names no resource of the Nudr_DR API.
// TS 29.500 Table 5.2.7.2-1 requires a 404 ProblemDetails with cause
// RESOURCE_URI_STRUCTURE_NOT_FOUND and TS 29.501 clause 4.8.2 requires the
// "application/problem+json" Content-Type, instead of gin's text/plain "404 page not found".
// The handlers return before touching the processor, so no database is needed.
func TestUDR_InfluenceDataUnmatchedSubResource_ResourceUriStructureNotFound(t *testing.T) {
	reqUri := factory.UdrDrResUriPrefix + "/application-data/influenceData/influenceId0001/no-such-resource"

	for _, method := range []string{http.MethodDelete, http.MethodGet, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			server := setupHttpServer(t)
			req, err := http.NewRequestWithContext(context.Background(), method, reqUri, nil)
			require.Nil(t, err)
			rsp := httptest.NewRecorder()
			server.ServeHTTP(rsp, req)

			require.Equal(t, http.StatusNotFound, rsp.Code)
			require.Equal(t, "application/problem+json", rsp.Header().Get("Content-Type"))
			require.NotContains(t, rsp.Body.String(), "404 page not found")

			problemDetail := models.ProblemDetails{}
			require.Nil(t, json.Unmarshal(rsp.Body.Bytes(), &problemDetail))
			require.Equal(t, int32(http.StatusNotFound), problemDetail.Status)
			require.Equal(t, "RESOURCE_URI_STRUCTURE_NOT_FOUND", problemDetail.Cause)
			require.Contains(t, problemDetail.Detail, reqUri)
		})
	}
}
