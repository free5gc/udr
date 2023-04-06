package datarepository_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/internal/sbi/datarepository"
	"github.com/free5gc/udr/internal/sbi/producer"
	"github.com/free5gc/udr/pkg/factory"
	util_logger "github.com/free5gc/util/logger"
	"github.com/free5gc/util/mongoapi"
)

func setupHttpServer() *gin.Engine {
	ginEngine := util_logger.NewGinWithLogrus(logger.GinLog)
	datarepository.AddService(ginEngine)
	return ginEngine
}

func setupMongoDB(t *testing.T) {
	err := mongoapi.SetMongoDB("test5gc", "mongodb://localhost:27017")
	require.Nil(t, err)
	err = mongoapi.Drop(producer.APPDATA_INFLUDATA_DB_COLLECTION_NAME)
	require.Nil(t, err)
	err = mongoapi.Drop(producer.APPDATA_INFLUDATA_SUBSC_DB_COLLECTION_NAME)
	require.Nil(t, err)
	err = mongoapi.Drop(producer.APPDATA_PFD_DB_COLLECTION_NAME)
	require.Nil(t, err)
}

func TestUDR_Root(t *testing.T) {
	server := setupHttpServer()
	reqUri := factory.UdrDrResUriPrefix + "/"

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqUri, nil)
	require.Nil(t, err)
	rsp := httptest.NewRecorder()
	server.ServeHTTP(rsp, req)

	t.Run("UDR Root", func(t *testing.T) {
		require.Equal(t, http.StatusOK, rsp.Code)
		require.Equal(t, "Hello World!", rsp.Body.String())
	})
}
