package util

import (
	"fmt"
	"free5gc/lib/openapi/models"
	udr_context "free5gc/src/udr/context"
	"free5gc/src/udr/factory"
	"free5gc/src/udr/logger"
	"os"

	"github.com/google/uuid"
)

func InitUdrContext(context *udr_context.UDRContext) {
	config := factory.UdrConfig
	logger.UtilLog.Infof("udrconfig Info: Version[%s] Description[%s]", config.Info.Version, config.Info.Description)
	configuration := config.Configuration
	context.NfId = uuid.New().String()
	sbi := configuration.Sbi
	context.UriScheme = models.UriScheme(sbi.Scheme)
	context.HttpIPv4Address = "127.0.0.1" // default localhost
	context.HttpIpv4Port = 29504          // default port
	if sbi != nil {
		if sbi.RegisterIPv4 != "" {
			context.HttpIPv4Address = sbi.RegisterIPv4
		}
		if sbi.Port != 0 {
			context.HttpIpv4Port = sbi.Port
		}
		context.BindingIPv4 = os.Getenv(sbi.BindingIPv4)
		if context.BindingIPv4 == "" {
			logger.UtilLog.Info("Problem parsing ServerIPv4 address from ENV Variable. Trying to parse it as string.")
			context.BindingIPv4 = sbi.BindingIPv4
			if context.BindingIPv4 == "" {
				logger.UtilLog.Info("Error parsing ServerIPv4 address as string. Using the localhost address as default.")
				context.BindingIPv4 = "0.0.0.0"
			}
		}
	}
	if configuration.NrfUri != "" {
		context.NrfUri = configuration.NrfUri
	} else {
		logger.UtilLog.Info("NRF Uri is empty! Using localhost as NRF IPv4 address.")
		context.NrfUri = fmt.Sprintf("%s://%s:%d", context.UriScheme, "127.0.0.1", 29510)	}
}
