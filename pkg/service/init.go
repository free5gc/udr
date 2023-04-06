package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/sirupsen/logrus"

	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/internal/sbi/consumer"
	"github.com/free5gc/udr/internal/sbi/datarepository"
	"github.com/free5gc/udr/pkg/factory"
	"github.com/free5gc/util/httpwrapper"
	logger_util "github.com/free5gc/util/logger"
	"github.com/free5gc/util/mongoapi"
)

type UdrApp struct {
	cfg    *factory.Config
	udrCtx *udr_context.UDRContext
}

func NewApp(cfg *factory.Config) (*UdrApp, error) {
	udr := &UdrApp{cfg: cfg}
	udr.SetLogEnable(cfg.GetLogEnable())
	udr.SetLogLevel(cfg.GetLogLevel())
	udr.SetReportCaller(cfg.GetLogReportCaller())
	udr_context.Init()
	udr.udrCtx = udr_context.GetSelf()
	return udr, nil
}

func (a *UdrApp) SetLogEnable(enable bool) {
	logger.MainLog.Infof("Log enable is set to [%v]", enable)
	if enable && logger.Log.Out == os.Stderr {
		return
	} else if !enable && logger.Log.Out == ioutil.Discard {
		return
	}
	a.cfg.SetLogEnable(enable)
	if enable {
		logger.Log.SetOutput(os.Stderr)
	} else {
		logger.Log.SetOutput(ioutil.Discard)
	}
}

func (a *UdrApp) SetLogLevel(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logger.MainLog.Warnf("Log level [%s] is invalid", level)
		return
	}
	logger.MainLog.Infof("Log level is set to [%s]", level)
	if lvl == logger.Log.GetLevel() {
		return
	}
	a.cfg.SetLogLevel(level)
	logger.Log.SetLevel(lvl)
}

func (a *UdrApp) SetReportCaller(reportCaller bool) {
	logger.MainLog.Infof("Report Caller is set to [%v]", reportCaller)
	if reportCaller == logger.Log.ReportCaller {
		return
	}
	a.cfg.SetLogReportCaller(reportCaller)
	logger.Log.SetReportCaller(reportCaller)
}

func (a *UdrApp) Start(tlsKeyLogPath string) {
	// get config file info
	config := factory.UdrConfig
	mongodb := config.Configuration.Mongodb

	logger.InitLog.Infof("UDR Config Info: Version[%s] Description[%s]", config.Info.Version, config.Info.Description)

	// Connect to MongoDB
	if err := mongoapi.SetMongoDB(mongodb.Name, mongodb.Url); err != nil {
		logger.InitLog.Errorf("UDR start err: %+v", err)
		return
	}

	logger.InitLog.Infoln("Server started")

	router := logger_util.NewGinWithLogrus(logger.GinLog)

	datarepository.AddService(router)

	pemPath := factory.UdrDefaultCertPemPath
	keyPath := factory.UdrDefaultPrivateKeyPath
	sbi := config.Configuration.Sbi
	if sbi.Tls != nil {
		pemPath = sbi.Tls.Pem
		keyPath = sbi.Tls.Key
	}

	self := a.udrCtx
	udr_context.InitUdrContext(self)

	addr := fmt.Sprintf("%s:%d", self.BindingIPv4, self.SBIPort)
	profile := consumer.BuildNFInstance(self)
	var newNrfUri string
	var err error
	newNrfUri, self.NfId, err = consumer.SendRegisterNFInstance(self.NrfUri, profile.NfInstanceId, profile)
	if err == nil {
		self.NrfUri = newNrfUri
	} else {
		logger.InitLog.Errorf("Send Register NFInstance Error[%s]", err.Error())
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				// Print stack for panic to log. Fatalf() will let program exit.
				logger.InitLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
			}
		}()

		<-signalChannel
		a.Terminate()
		os.Exit(0)
	}()

	server, err := httpwrapper.NewHttp2Server(addr, tlsKeyLogPath, router)
	if server == nil {
		logger.InitLog.Errorf("Initialize HTTP server failed: %+v", err)
		return
	}

	if err != nil {
		logger.InitLog.Warnf("Initialize HTTP server: %+v", err)
	}

	serverScheme := factory.UdrConfig.Configuration.Sbi.Scheme
	if serverScheme == "http" {
		err = server.ListenAndServe()
	} else if serverScheme == "https" {
		err = server.ListenAndServeTLS(pemPath, keyPath)
	}

	if err != nil {
		logger.InitLog.Fatalf("HTTP server setup failed: %+v", err)
	}
}

func (a *UdrApp) Terminate() {
	logger.InitLog.Infof("Terminating UDR...")
	// deregister with NRF
	problemDetails, err := consumer.SendDeregisterNFInstance()
	if problemDetails != nil {
		logger.InitLog.Errorf("Deregister NF instance Failed Problem[%+v]", problemDetails)
	} else if err != nil {
		logger.InitLog.Errorf("Deregister NF instance Error[%+v]", err)
	} else {
		logger.InitLog.Infof("Deregister from NRF successfully")
	}
	logger.InitLog.Infof("UDR terminated")
}
