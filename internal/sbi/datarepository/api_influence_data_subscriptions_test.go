package datarepository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/udr/pkg/factory"
)

func TestUDR_GetSubs2Notify_GetBeforeCreateingOne(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	server := setupHttpServer()
	reqUri := factory.UdrDrResUriPrefix + "/application-data/influenceData/subs-to-notify?dnn=internet"

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqUri, nil)
	require.Nil(t, err)
	rsp := httptest.NewRecorder()
	server.ServeHTTP(rsp, req)

	t.Run("UDR subs-to-notify Get before Create, dnn==internet", func(t *testing.T) {
		require.Equal(t, http.StatusOK, rsp.Code)
		require.Equal(t, "null", rsp.Body.String())
	})
}

func TestUDR_GetSubs2Notify_CreateThenGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	server := setupHttpServer()
	baseUri := factory.UdrDrResUriPrefix + "/application-data/influenceData/subs-to-notify"
	reqUri := baseUri

	test := models.TrafficInfluSub{
		Dnns: []string{"internet", "outernet"},
		Snssais: []models.Snssai{{
			Sst: 1, Sd: "010203",
		}, {
			Sst: 1, Sd: "112233",
		}},
		NotificationUri: "http://127.0.0.1/notifyMePlease",
	}
	bjson, err := json.Marshal(test)
	require.Nil(t, err)
	reqBody := bytes.NewReader(bjson)

	test.NotificationUri = ""
	bjson_bad, err := json.Marshal(test)
	require.Nil(t, err)
	reqBody_bad := bytes.NewReader(bjson_bad)

	// Create one - w/o the mandatory NotificationUri:
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqUri, reqBody_bad)
	require.Nil(t, err)
	rsp := httptest.NewRecorder()
	server.ServeHTTP(rsp, req)
	t.Run("UDR subs-to-notify CreateThenGet - Create w/o mandatory notificationUri", func(t *testing.T) {
		require.Equal(t, http.StatusBadRequest, rsp.Code)
	})

	// Create one - normal
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, reqUri, reqBody)
	require.Nil(t, err)
	rsp = httptest.NewRecorder()
	server.ServeHTTP(rsp, req)
	// Linter complains not closing response body even tried to close,
	// ==> remove the comments manually to test if location header is there.
	// location := rsp.Result().Header.Get("Location")
	// err = rsp.Result().Body.Close()
	// require.Nil(t, err)
	// require.NotNil(t, location)
	// t.Log("location:", location)
	t.Run("UDR subs-to-notify CreateThenGet - Create normal case", func(t *testing.T) {
		require.Equal(t, http.StatusCreated, rsp.Code)
		require.Equal(t, string(bjson), rsp.Body.String())
		// require.True(t, strings.Contains(location, baseUri+"/"))
		// require.True(t, strings.HasPrefix(location, udr_context.GetSelf().GetIPv4Uri()+baseUri+"/"))
	})

	// Get success
	rsp = getUri(t, baseUri, "?dnn=internet")
	t.Run("UDR subs-to-notify CreateThenGet - get", func(t *testing.T) {
		require.Equal(t, http.StatusOK, rsp.Code)
		require.Equal(t, "["+string(bjson)+"]", rsp.Body.String())
	})

	// Get without a filter
	rsp = getUri(t, baseUri, "")
	t.Run("UDR subs-to-notify CreateThenGet - get w/o a filter", func(t *testing.T) {
		require.Equal(t, http.StatusBadRequest, rsp.Code)
	})

	// Get a non-exist DNN
	rsp = getUri(t, baseUri, "?dnn=ThisIsABadDNN")
	t.Run("UDR subs-to-notify CreateThenGet - get bad DNN", func(t *testing.T) {
		require.Equal(t, http.StatusOK, rsp.Code)
		require.Equal(t, "null", rsp.Body.String())
	})
}
