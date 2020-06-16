package util

import (
	"fmt"
	"os"
	"free5gc/lib/openapi/models"
	udr_context "free5gc/src/udr/context"
	"free5gc/src/udr/factory"
	"free5gc/src/udr/logger"

	"github.com/google/uuid"
)

func InitUdrContext(context *udr_context.UDRContext) {
	config := factory.UdrConfig
	logger.UtilLog.Infof("udrconfig Info: Version[%s] Description[%s]", config.Info.Version, config.Info.Description)
	configuration := config.Configuration
	context.NfId = uuid.New().String()
	context.ServerIPv4 = os.Getenv(configuration.ServerIPv4)
	if context.ServerIPv4 == "" {
		logger.UtilLog.Warn("Problem parsing ServerIPv4 address from ENV Variable. Trying to parse it as string.")
		context.ServerIPv4 = configuration.ServerIPv4
		if context.ServerIPv4 == "" {
			logger.UtilLog.Warn("Error parsing ServerIPv4 address as string. Using the localhost address as default.")
			context.ServerIPv4 = "127.0.0.1"
		}
	}
	sbi := configuration.Sbi
	context.UriScheme = models.UriScheme(sbi.Scheme)
	context.HttpIPv4Address = "127.0.0.1" // default localhost
	context.HttpIpv4Port = 29504          // default port
	if sbi != nil {
		if sbi.IPv4Addr != "" {
			context.HttpIPv4Address = sbi.IPv4Addr
		}
		if sbi.Port != 0 {
			context.HttpIpv4Port = sbi.Port
		}
	}
	if configuration.NrfUri != "" {
		context.NrfUri = configuration.NrfUri
	} else {
		logger.UtilLog.Warn("NRF Uri is empty! Using localhost as NRF IPv4 address.")
		context.NrfUri = fmt.Sprintf("%s://%s:%d", context.UriScheme, "127.0.0.1", 29510)
	}
}
