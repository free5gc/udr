package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/openapi/models"
	NFManagement "github.com/free5gc/openapi/nrf/NFMgmt"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	udr_util "github.com/free5gc/udr/internal/util"
	"github.com/free5gc/util/metrics/sbi"
)

func callbackServiceNameForNfType(
	nfType models.Nrf_NFMgmt_NFType,
	defaultServiceName models.Nrf_NFMgmt_ServiceName,
) models.Nrf_NFMgmt_ServiceName {
	switch nfType {
	case models.Nrf_NFMgmt_NFType_UDM:
		return models.Nrf_NFMgmt_ServiceName_NUDM_SDM
	case models.Nrf_NFMgmt_NFType_PCF:
		return defaultServiceName
	case models.Nrf_NFMgmt_NFType_NEF:
		return models.Nrf_NFMgmt_ServiceName_NNEF_EVENTEXPOSURE
	default:
		return defaultServiceName
	}
}

func requesterNFTypeFromNRF(nfInstanceID string) (models.Nrf_NFMgmt_NFType, bool) {
	if nfInstanceID == "" {
		return "", false
	}

	udrSelf := udr_context.GetSelf()
	if !udrSelf.OAuth2Required {
		return "", false
	}

	ctx, pd, err := udrSelf.GetTokenCtxForNRF(models.Nrf_NFMgmt_ServiceName_NNRF_NFM)
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
	if rsp == nil || rsp.Nrf_NFMgmt_NFProfile == nil || rsp.Nrf_NFMgmt_NFProfile.NfType == "" {
		logger.SBILog.Errorf("Get requester NF profile [%s] returned empty NF type", nfInstanceID)
		return "", false
	}
	return rsp.Nrf_NFMgmt_NFProfile.NfType, true
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
	defaultServiceName models.Nrf_NFMgmt_ServiceName, defaultNfType models.Nrf_NFMgmt_NFType,
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
