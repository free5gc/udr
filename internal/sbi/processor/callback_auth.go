package processor

import (
	"context"
	"net/url"
	"strings"

	"github.com/free5gc/openapi/models"
	udr_context "github.com/free5gc/udr/internal/context"
)

func callbackRequestContext(callbackURI string) (context.Context, *models.ProblemDetails, error) {
	serviceName, targetNF, ok := callbackServiceInfo(callbackURI)
	if !ok {
		return context.TODO(), nil, nil
	}
	return udr_context.GetSelf().GetTokenCtx(serviceName, targetNF)
}

func callbackServiceInfo(callbackURI string) (models.ServiceName, models.NrfNfManagementNfType, bool) {
	parsedURI, err := url.Parse(callbackURI)
	if err != nil {
		return "", "", false
	}

	path := strings.TrimPrefix(parsedURI.Path, "/")
	if path == "" {
		return "", "", false
	}

	serviceName := strings.SplitN(path, "/", 2)[0]
	switch serviceName {
	case "namf-callback":
		return models.ServiceName(serviceName), models.NrfNfManagementNfType_AMF, true
	case "nausf-callback":
		return models.ServiceName(serviceName), models.NrfNfManagementNfType_AUSF, true
	case "nsmf-callback":
		return models.ServiceName(serviceName), models.NrfNfManagementNfType_SMF, true
	case "nudm-callback":
		return models.ServiceName(serviceName), models.NrfNfManagementNfType_UDM, true
	case "npcf-callback":
		return models.ServiceName(serviceName), models.NrfNfManagementNfType_PCF, true
	case "nnef-callback":
		return models.ServiceName(serviceName), models.NrfNfManagementNfType_NEF, true
	case "nchf-callback":
		return models.ServiceName(serviceName), models.NrfNfManagementNfType_CHF, true
	case "nbsf-callback":
		return models.ServiceName(serviceName), models.NrfNfManagementNfType_BSF, true
	case "nnssf-callback":
		return models.ServiceName(serviceName), models.NrfNfManagementNfType_NSSF, true
	case "nudr-callback":
		return models.ServiceName(serviceName), models.NrfNfManagementNfType_UDR, true
	default:
		pathParts := strings.Split(path, "/")
		if len(pathParts) == 2 && pathParts[1] == "sdm-subscriptions" {
			return models.ServiceName_NUDM_SDM, models.NrfNfManagementNfType_UDM, true
		}
		return "", "", false
	}
}
