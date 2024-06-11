package service

import (
	"context"
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
	"github.com/free5gc/udr/internal/sbi/processor"
	"github.com/free5gc/udr/pkg/app"
	"github.com/free5gc/udr/pkg/factory"
	"github.com/free5gc/util/mongoapi"
)

type UdrApp struct {
	cfg    *factory.Config
	udrCtx *udr_context.UDRContext

	wg        sync.WaitGroup
	sbiServer *sbi.Server
	processor *processor.Processor
	consumer  *consumer.Consumer
}

var _ app.UdrApp = &UdrApp{}

func NewApp(cfg *factory.Config, tlsKeyLogPath string) (*UdrApp, error) {
	udr_context.Init()
	udr_context.InitUdrContext()
	udr := &UdrApp{
		cfg:    cfg,
		udrCtx: udr_context.GetSelf(),
		wg:     sync.WaitGroup{},
	}
	udr.SetLogEnable(cfg.GetLogEnable())
	udr.SetLogLevel(cfg.GetLogLevel())
	udr.SetReportCaller(cfg.GetLogReportCaller())

	processor := processor.NewProcessor(udr)
	udr.processor = processor

	consumer := consumer.NewConsumer(udr)
	udr.consumer = consumer

	udr.sbiServer = sbi.NewServer(udr, tlsKeyLogPath)

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

func (u *UdrApp) registerToNrf(ctx context.Context) error {
	udrContext := u.udrCtx

	nrfUri, nfId, err := u.consumer.SendRegisterNFInstance(ctx, udrContext.NrfUri)
	if err != nil {
		return fmt.Errorf("send register NFInstance error[%s]", err.Error())
	}
	udrContext.NrfUri = nrfUri
	udrContext.NfId = nfId

	return nil
}

func (u *UdrApp) deregisterFromNrf() {
	problemDetails, err := u.consumer.SendDeregisterNFInstance()
	if problemDetails != nil {
		logger.InitLog.Errorf("Deregister NF instance Failed Problem[%+v]", problemDetails)
	} else if err != nil {
		logger.InitLog.Errorf("Deregister NF instance Error[%+v]", err)
	} else {
		logger.InitLog.Infof("Deregister from NRF successfully")
	}
}

func (a *UdrApp) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh  // Wait for interrupt signal to gracefully shutdown
		cancel() // Notify each goroutine and wait them stopped
	}()

	err := a.registerToNrf(ctx)
	if err != nil {
		logger.InitLog.Errorf("register to NRF failed: %v", err)
	} else {
		logger.InitLog.Infof("register to NRF successfully")
	}

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

	// Graceful deregister when panic
	defer func() {
		if p := recover(); p != nil {
			logger.InitLog.Errorf("panic: %v\n%s", p, string(debug.Stack()))
			a.deregisterFromNrf()
		}
	}()

	a.sbiServer.Run(&a.wg)
	go a.listenShutdown(ctx)
}

func (a *UdrApp) Processor() *processor.Processor {
	return a.processor
}

func (a *UdrApp) Consumer() *consumer.Consumer {
	return a.consumer
}

func (a *UdrApp) listenShutdown(ctx context.Context) {
	<-ctx.Done()
	a.Terminate()
}

func (a *UdrApp) Terminate() {
	logger.InitLog.Infof("Terminating UDR...")
	// deregister with NRF
	a.deregisterFromNrf()
	a.sbiServer.Shutdown()
	logger.InitLog.Infof("UDR terminated")
}

func (a *UdrApp) Wait() {
	a.wg.Wait()
	logger.MainLog.Infof("UDR terminated")
}
