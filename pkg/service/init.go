package service

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/internal/sbi"
	"github.com/free5gc/udr/internal/sbi/consumer"
	"github.com/free5gc/udr/pkg/factory"
	"github.com/free5gc/util/mongoapi"
)

type UdrApp struct {
	cfg    *factory.Config
	udrCtx *udr_context.UDRContext

	sbiServer *sbi.Server
	wg 	  sync.WaitGroup
}

func NewApp(cfg *factory.Config, tlsKeyLogPath string) (*UdrApp, error) {
	udr := &UdrApp{
		cfg: cfg,
		wg: sync.WaitGroup{},
	}
	udr.SetLogEnable(cfg.GetLogEnable())
	udr.SetLogLevel(cfg.GetLogLevel())
	udr.SetReportCaller(cfg.GetLogReportCaller())

	udr.sbiServer = sbi.NewServer(udr, tlsKeyLogPath)
	udr_context.Init()
	udr.udrCtx = udr_context.GetSelf()
	return udr, nil
}

func (u *UdrApp) Config() *factory.Config {
	return u.cfg
}

func (u *UdrApp) Context() *udr_context.UDRContext {
	return u.udrCtx
}

func (a *UdrApp) SetLogEnable(enable bool) {
	logger.MainLog.Infof("Log enable is set to [%v]", enable)
	if enable && logger.Log.Out == os.Stderr {
		return
	} else if !enable && logger.Log.Out == io.Discard {
		return
	}
	a.cfg.SetLogEnable(enable)
	if enable {
		logger.Log.SetOutput(os.Stderr)
	} else {
		logger.Log.SetOutput(io.Discard)
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

func (u *UdrApp) registerToNrf() error {
	udrContext := u.udrCtx
	profile, err := consumer.BuildNFInstance(udrContext)
	if err != nil {
		return fmt.Errorf("Build NF Instance Error[%s]", err.Error())
	}

	udrContext.NrfUri, udrContext.NfId, err = consumer.SendRegisterNFInstance(udrContext.NrfUri, profile.NfInstanceId, profile)
	if err != nil {
		return fmt.Errorf("Send Register NFInstance Error[%s]", err.Error())
	}
	return nil
}


func (u *UdrApp) deregisterFromNrf() {
	problemDetails, err := consumer.SendDeregisterNFInstance()
	if problemDetails != nil {
		logger.InitLog.Errorf("Deregister NF instance Failed Problem[%+v]", problemDetails)
	} else if err != nil {
		logger.InitLog.Errorf("Deregister NF instance Error[%+v]", err)
	} else {
		logger.InitLog.Infof("Deregister from NRF successfully")
	}
}

func (a *UdrApp) addSigTermHandler() {
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
} 

func (a *UdrApp) Start(tlsKeyLogPath string) {
	// get config file info
	logger.InitLog.Infoln("Server started")
	config := factory.UdrConfig
	mongodb := config.Configuration.Mongodb

	logger.InitLog.Infof("UDR Config Info: Version[%s] Description[%s]", config.Info.Version, config.Info.Description)

	// Connect to MongoDB
	if err := mongoapi.SetMongoDB(mongodb.Name, mongodb.Url); err != nil {
		logger.InitLog.Errorf("UDR start set MongoDB error: %+v", err)
		return
	}

	err := a.registerToNrf()
	if err != nil {
		logger.InitLog.Errorf("Register to NRF failed: %+v", err)
	}
	// Graceful deregister when panic
	defer func() {
		if p := recover(); p != nil {
			logger.InitLog.Errorf("panic: %v\n%s", p, string(debug.Stack()))
			a.deregisterFromNrf()
		}
	}()

	a.sbiServer.Run(&a.wg)
	a.addSigTermHandler()
}

func (a *UdrApp) Terminate() {
	logger.InitLog.Infof("Terminating UDR...")
	// deregister with NRF
	a.deregisterFromNrf()
	a.sbiServer.Shutdown()
	logger.InitLog.Infof("UDR terminated")
}
