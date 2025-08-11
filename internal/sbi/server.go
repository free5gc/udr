package sbi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/internal/sbi/processor"
	"github.com/free5gc/udr/internal/util"
	"github.com/free5gc/udr/pkg/app"
	"github.com/free5gc/udr/pkg/factory"
	"github.com/free5gc/util/httpwrapper"
	logger_util "github.com/free5gc/util/logger"
	"github.com/free5gc/util/metrics"
)

type Server struct {
	UDR

	httpServer *http.Server
	router     *gin.Engine
}

type UDR interface {
	app.App

	Processor() *processor.Processor
}

func NewServer(udr UDR, tlsKeyLogPath string) *Server {
	s := &Server{
		UDR: udr,
	}


	s.router = newRouter(s)
	
	addr := s.Context().RegisterIP
	port := uint16(s.Context().SBIPort)

	bind := netip.AddrPortFrom(addr, port).String()
	bindAddr := fmt.Sprintf("%s", bind)

	logger.SBILog.Infof("Binding addr: [%s]", bindAddr)
	var err error
	if s.httpServer, err = httpwrapper.NewHttp2Server(bindAddr, tlsKeyLogPath, s.router); err != nil {
		logger.InitLog.Errorf("Initialize HTTP server failed: %v", err)
		panic("Server initialization failed")
	}
	s.httpServer.ErrorLog = log.New(logger.SBILog.WriterLevel(logrus.ErrorLevel), "HTTP2: ", 0)

	return s
}

func (s *Server) Run(wg *sync.WaitGroup) {
	logger.SBILog.Info("Starting server...")

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := s.serve()
		if err != http.ErrServerClosed {
			logger.SBILog.Panicf("HTTP server setup failed: %+v", err)
		}
		logger.SBILog.Infof("SBI server (listen on %s) stopped", s.httpServer.Addr)
	}()
}

func (s *Server) Shutdown() {
	s.shutdownHttpServer()
}

func (s *Server) shutdownHttpServer() {
	const shutdownTimeout time.Duration = 2 * time.Second

	if s.httpServer == nil {
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := s.httpServer.Shutdown(shutdownCtx)
	if err != nil {
		logger.SBILog.Errorf("HTTP server shutdown failed: %+v", err)
	}
}

func newRouter(s *Server) *gin.Engine {
	router := logger_util.NewGinWithLogrus(logger.GinLog)
	router.Use(metrics.InboundMetrics())

	dataRepositoryGroup := router.Group(factory.UdrDrResUriPrefix)
	dataRepositoryGroup.Use(func(c *gin.Context) {
		util.NewRouterAuthorizationCheck(models.ServiceName_NUDR_DR).Check(c, s.Context())
	})
	dataRepositoryRoutes := s.getDataRepositoryRoutes()
	AddService(dataRepositoryGroup, dataRepositoryRoutes)

	groupIdGroup := router.Group(factory.UdrGroupIdResUriPrefix)
	groupIdGroup.Use(func(c *gin.Context) {
		util.NewRouterAuthorizationCheck(models.ServiceName_NUDR_GROUP_ID_MAP).Check(c, s.Context())
	})
	groupIdRoutes := s.getGroupIdMap()
	AddService(groupIdGroup, groupIdRoutes)

	imsSDM := router.Group(factory.HSSIsmSDMUriPrefix)
	imsSDM.Use(func(c *gin.Context) {
		util.NewRouterAuthorizationCheck(models.ServiceName_NHSS_IMS_SDM).Check(c, s.Context())
	})
	imsSDMRoutes := s.getImsSDMRoutes()
	AddService(imsSDM, imsSDMRoutes)

	return router
}

func (s *Server) unsecureServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) secureServe() error {
	sbiConfig := s.UDR.Config()

	pemPath := sbiConfig.GetCertPemPath()
	if pemPath == "" {
		pemPath = factory.UdrDefaultCertPemPath
	}

	keyPath := sbiConfig.GetCertKeyPath()
	if keyPath == "" {
		keyPath = factory.UdrDefaultPrivateKeyPath
	}

	return s.httpServer.ListenAndServeTLS(pemPath, keyPath)
}

func (s *Server) serve() error {
	sbiConfig := s.UDR.Config().Configuration.Sbi

	switch sbiConfig.Scheme {
	case "http":
		return s.unsecureServe()
	case "https":
		return s.secureServe()
	default:
		return fmt.Errorf("invalid SBI scheme: %s", sbiConfig.Scheme)
	}
}
