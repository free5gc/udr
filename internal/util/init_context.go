package util

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"

	"github.com/google/uuid"

	"github.com/free5gc/openapi/models"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/pkg/factory"
)

func InitUdrContext(context *udr_context.UDRContext) {
	config := factory.UdrConfig
	logger.UtilLog.Infof("udrconfig Info: Version[%s] Description[%s]", config.Info.Version, config.Info.Description)
	configuration := config.Configuration
	context.NfId = uuid.New().String()

    sbi := configuration.Sbi

    context.SBIPort = sbi.Port

    context.UriScheme = models.UriScheme(sbi.Scheme)

    if bindingIP := os.Getenv(sbi.BindingIP); bindingIP != "" {
            logger.UtilLog.Info("Parsing BindingIP address from ENV Variable.")
            sbi.BindingIP = bindingIP
    }
    if registerIP := os.Getenv(sbi.RegisterIP); registerIP != "" {
            logger.UtilLog.Info("Parsing RegisterIP address from ENV Variable.")
            sbi.RegisterIP = registerIP
    }
    context.BindingIP = resolveIP(sbi.BindingIP)
    context.RegisterIP = resolveIP(sbi.RegisterIP)

	if configuration.NrfUri != "" {
		context.NrfUri = configuration.NrfUri
	} else {
		logger.UtilLog.Warn("NRF Uri is empty! Using localhost as NRF IPv4 address.")
		context.NrfUri = fmt.Sprintf("%s://%s:%d", context.UriScheme, "127.0.0.1", 29510)
	}
}

func resolveIP(ip string) netip.Addr {
	resolvedIPs, err := net.DefaultResolver.LookupNetIP(context.Background(), "ip", ip)
	if err != nil {
		logger.InitLog.Errorf("Lookup failed with %s: %+v", ip, err)
	}
	resolvedIP := resolvedIPs[0].Unmap()
	if resolvedIP := resolvedIP.String(); resolvedIP != ip {
		logger.UtilLog.Infof("Lookup revolved %s into %s", ip, resolvedIP)
	}
	return resolvedIP
}
