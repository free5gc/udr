package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi/nrf/NFManagement"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	udr_util "github.com/free5gc/udr/internal/util"
	"github.com/free5gc/util/metrics/sbi"
)

func callbackServiceNameForNfType(
	nfType models.NrfNfManagementNfType,
	defaultServiceName models.ServiceName,
) models.ServiceName {
	switch nfType {
	case models.NrfNfManagementNfType_UDM:
		return models.ServiceName_NUDM_SDM
	case models.NrfNfManagementNfType_PCF:
		return defaultServiceName
	case models.NrfNfManagementNfType_NEF:
		return models.ServiceName_NNEF_EVENTEXPOSURE
	default:
		return defaultServiceName
	}
}

func requesterNFTypeFromNRF(nfInstanceID string) (models.NrfNfManagementNfType, bool) {
	if nfInstanceID == "" {
		return "", false
	}

	udrSelf := udr_context.GetSelf()
	if !udrSelf.OAuth2Required {
		return "", false
	}

	ctx, pd, err := udrSelf.GetTokenCtx(models.ServiceName_NNRF_NFM, models.NrfNfManagementNfType_NRF)
	if err != nil {
		logger.SBILog.Errorf("Get requester NF profile token failed: %v", err)
		if pd != nil {
			logger.SBILog.Errorf("Get requester NF profile token problem details: %+v", pd)
		}
		return "", false
	}
	if pd != nil {
		logger.SBILog.Errorf("Get requester NF profile token problem details: %+v", pd)
		return "", false
	}

	configuration := NFManagement.NewConfiguration()
	configuration.SetBasePath(udrSelf.NrfUri)
	client := NFManagement.NewAPIClient(configuration)
	request := &NFManagement.GetNFInstanceRequest{}
	request.SetNfInstanceID(nfInstanceID)

	rsp, err := client.NFInstanceIDDocumentApi.GetNFInstance(ctx, request)
	if err != nil {
		logger.SBILog.Errorf("Get requester NF profile [%s] failed: %v", nfInstanceID, err)
		return "", false
	}
	if rsp == nil || rsp.NrfNfManagementNfProfile.NfType == "" {
		logger.SBILog.Errorf("Get requester NF profile [%s] returned empty NF type", nfInstanceID)
		return "", false
	}
	return rsp.NrfNfManagementNfProfile.NfType, true
}

func rejectUnresolvedCallbackTarget(c *gin.Context) {
	pd := models.ProblemDetails{
		Status: http.StatusUnauthorized,
		Cause:  "REQUESTER_IDENTITY_UNRESOLVED",
		Detail: "Unable to resolve requester NF identity for callback token target",
	}
	c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
	c.JSON(int(pd.Status), pd)
}

func subscriptionCallbackTargetFromContext(
	c *gin.Context,
	defaultServiceName models.ServiceName,
	defaultNfType models.NrfNfManagementNfType,
) (udr_context.SubscriptionCallbackTarget, bool) {
	target := udr_context.SubscriptionCallbackTarget{
		ServiceName: defaultServiceName,
		NfType:      defaultNfType,
	}

	if !udr_context.GetSelf().OAuth2Required {
		return target, true
	}

	if nfInstanceID, ok := c.Get(udr_util.RequesterNfInstanceIDCtxKey); ok {
		if nfInstanceID, ok := nfInstanceID.(string); ok {
			target.NfInstanceID = nfInstanceID
		}
	}

	if nfType, ok := requesterNFTypeFromNRF(target.NfInstanceID); ok {
		target.NfType = nfType
		target.ServiceName = callbackServiceNameForNfType(nfType, defaultServiceName)
		return target, true
	}

	return target, false
}
