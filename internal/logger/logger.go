package logger

import (
	"strings"

	"github.com/sirupsen/logrus"

	logger_util "github.com/free5gc/util/logger"
)

var (
	Log         *logrus.Logger
	NfLog       *logrus.Entry
	MainLog     *logrus.Entry
	InitLog     *logrus.Entry
	CfgLog      *logrus.Entry
	CtxLog      *logrus.Entry
	DataRepoLog *logrus.Entry
	UtilLog     *logrus.Entry
	HttpLog     *logrus.Entry
	ConsumerLog *logrus.Entry
	GinLog      *logrus.Entry
	ProcLog     *logrus.Entry
	SBILog      *logrus.Entry
	DbLog       *logrus.Entry
)

// ginLogSanitizer prevents log forging through untrusted HTTP request data.
type ginLogSanitizer struct{}

// Levels enables the sanitizer for all log levels.
func (ginLogSanitizer) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire escapes newline characters in Gin access logs before they are written.
func (ginLogSanitizer) Fire(entry *logrus.Entry) error {
	// Only sanitize Gin access logs to avoid changing other log categories.
	if entry.Data[logger_util.FieldCategory] != "GIN" {
		return nil
	}

	// Escape CR and LF to prevent injected content from creating fake log lines.
	entry.Message = strings.NewReplacer(
		"\r", `\r`,
		"\n", `\n`,
	).Replace(entry.Message)

	return nil
}

func init() {
	fieldsOrder := []string{
		logger_util.FieldNF,
		logger_util.FieldCategory,
	}

	Log = logger_util.New(fieldsOrder)
	// Register the sanitizer to protect Gin access logs from log forging.
	Log.AddHook(ginLogSanitizer{})
	NfLog = Log.WithField(logger_util.FieldNF, "UDR")
	MainLog = NfLog.WithField(logger_util.FieldCategory, "Main")
	InitLog = NfLog.WithField(logger_util.FieldCategory, "Init")
	CfgLog = NfLog.WithField(logger_util.FieldCategory, "CFG")
	CtxLog = NfLog.WithField(logger_util.FieldCategory, "CTX")
	GinLog = NfLog.WithField(logger_util.FieldCategory, "GIN")
	ConsumerLog = NfLog.WithField(logger_util.FieldCategory, "Consumer")
	DataRepoLog = NfLog.WithField(logger_util.FieldCategory, "DataRepo")
	ProcLog = NfLog.WithField(logger_util.FieldCategory, "Proc")
	HttpLog = NfLog.WithField(logger_util.FieldCategory, "HTTP")
	UtilLog = NfLog.WithField(logger_util.FieldCategory, "Util")
	SBILog = NfLog.WithField(logger_util.FieldCategory, "SBI")
	DbLog = NfLog.WithField(logger_util.FieldCategory, "DB")
}
