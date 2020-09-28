package logger

import (
	"os"
	"time"

	formatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"

	"free5gc/lib/logger_conf"
	"free5gc/lib/logger_util"
)

var log *logrus.Logger
var AppLog *logrus.Entry
var InitLog *logrus.Entry
var HandlerLog *logrus.Entry
var DataRepoLog *logrus.Entry
var UtilLog *logrus.Entry
var HttpLog *logrus.Entry
var GinLog *logrus.Entry

func init() {
	log = logrus.New()
	log.SetReportCaller(false)

	log.Formatter = &formatter.Formatter{
		TimestampFormat: time.RFC3339,
		TrimMessages:    true,
		NoFieldsSpace:   true,
		HideKeys:        true,
		FieldsOrder:     []string{"component", "category"},
	}

	free5gcLogHook, err := logger_util.NewFileHook(logger_conf.Free5gcLogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err == nil {
		log.Hooks.Add(free5gcLogHook)
	}

	selfLogHook, err := logger_util.NewFileHook(logger_conf.NfLogDir+"udr.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err == nil {
		log.Hooks.Add(selfLogHook)
	}

	AppLog = log.WithFields(logrus.Fields{"component": "UDR", "category": "App"})
	InitLog = log.WithFields(logrus.Fields{"component": "UDR", "category": "Init"})
	HandlerLog = log.WithFields(logrus.Fields{"component": "UDR", "category": "Handler"})
	DataRepoLog = log.WithFields(logrus.Fields{"component": "UDR", "category": "DataRepo"})
	UtilLog = log.WithFields(logrus.Fields{"component": "UDR", "category": "Util"})
	HttpLog = log.WithFields(logrus.Fields{"component": "UDR", "category": "HTTP"})
	GinLog = log.WithFields(logrus.Fields{"component": "UDR", "category": "GIN"})
}

func SetLogLevel(level logrus.Level) {
	log.SetLevel(level)
}

func SetReportCaller(bool bool) {
	log.SetReportCaller(bool)
}
