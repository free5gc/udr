package consumer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	NFDiscovery "github.com/free5gc/openapi/nrf/NFDisc"
	NFManagement "github.com/free5gc/openapi/nrf/NFMgmt"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	sbi_metrics "github.com/free5gc/util/metrics/sbi"
	"github.com/free5gc/util/nfheartbeat"
)

const registerRetryInterval = 2 * time.Second

type NrfService struct {
	nfMngmntMu sync.RWMutex

	nfMngmntClients map[string]*NFManagement.APIClient

	heartbeat *nfheartbeat.Runner

	// Interval in seconds from the last registration response. Written before the heartbeat starts, then only on
	// the heartbeat goroutine.
	heartbeatTimer int32
}

func (ns *NrfService) getNFManagementClient(uri string) *NFManagement.APIClient {
	if uri == "" {
		return nil
	}
	ns.nfMngmntMu.RLock()
	client, ok := ns.nfMngmntClients[uri]
	ns.nfMngmntMu.RUnlock()
	if ok {
		return client
	}

	ns.nfMngmntMu.Lock()
	defer ns.nfMngmntMu.Unlock()
	// Another caller may have stored a client while the read lock was down.
	if client, ok = ns.nfMngmntClients[uri]; ok {
		return client
	}

	configuration := NFManagement.NewConfiguration()
	configuration.SetBasePath(uri)
	configuration.SetMetrics(sbi_metrics.SbiMetricHook)
	client = NFManagement.NewAPIClient(configuration)

	ns.nfMngmntClients[uri] = client
	return client
}

func (ns *NrfService) buildNFProfile(context *udr_context.UDRContext) (models.Nrf_NFMgmt_NFProfile, error) {
	// config := factory.UdrConfig

	profile := models.Nrf_NFMgmt_NFProfile{
		NfInstanceId:  context.NfId,
		NfType:        models.Nrf_NFMgmt_NFType_UDR,
		NfStatus:      models.Nrf_NFMgmt_NFStatus_REGISTERED,
		Ipv4Addresses: []string{context.RegisterIPv4},
		UdrInfo: &models.Nrf_NFMgmt_UdrInfo{
			SupportedDataSets: []models.Nrf_NFMgmt_DataSetId{
				models.Nrf_NFMgmt_DataSetId_SUBSCRIPTION,
			},
		},
	}

	var services []models.Nrf_NFMgmt_NFService
	for _, nfService := range context.NfService {
		services = append(services, nfService)
	}
	if len(services) > 0 {
		profile.NfServices = services
	}

	return profile, nil
}

// SendRegisterNFInstance PUTs the profile on the NF's own instance ID (3GPP TS 29.510 clause 5.2.2.2.2) until it
// succeeds or ctx is done. Only the startup call may pass applyOAuth2: it writes OAuth2Required unsynchronized.
func (ns *NrfService) SendRegisterNFInstance(ctx context.Context, applyOAuth2 bool) error {
	udrContext := udr_context.GetSelf()
	client := ns.getNFManagementClient(udrContext.NrfUri)
	if client == nil {
		return fmt.Errorf("SendRegisterNFInstance: nrf not found")
	}

	profile, err := ns.buildNFProfile(udrContext)
	if err != nil {
		return fmt.Errorf("failed to build nrf profile %s", err.Error())
	}

	var rsp *NFManagement.RegisterNFInstanceResponse
	registerReq := &NFManagement.RegisterNFInstanceRequest{
		NfInstanceID: &profile.NfInstanceId,
		RequestBody:  &profile,
	}
	for ctx.Err() == nil {
		rsp, err = client.NFInstanceIDDocumentApi.RegisterNFInstance(ctx, registerReq)
		if err == nil && rsp != nil {
			var nf models.Nrf_NFMgmt_NFProfile
			if rsp.Nrf_NFMgmt_NFProfile != nil {
				nf = *rsp.Nrf_NFMgmt_NFProfile
			}
			ns.processRegisterResponse(udrContext, nf, applyOAuth2)
			return nil
		}
		logger.ConsumerLog.Errorf("UDR register to NRF Error[%v]", err)
		select {
		case <-ctx.Done():
		case <-time.After(registerRetryInterval):
		}
	}
	return fmt.Errorf("NFRegister aborted: %w (last error: %v)", ctx.Err(), err)
}

// processRegisterResponse adopts what the NRF answered to the NFRegister PUT: the
// heartbeat interval and the oauth2 custom info.
func (ns *NrfService) processRegisterResponse(
	udrContext *udr_context.UDRContext,
	nf models.Nrf_NFMgmt_NFProfile,
	applyOAuth2 bool,
) {
	ns.heartbeatTimer = nf.HeartBeatTimer

	oauth2 := false
	if customInfo, ok := nf.CustomInfo.(map[string]interface{}); ok {
		if v, isBool := customInfo["oauth2"].(bool); isBool {
			oauth2 = v
			logger.MainLog.Infoln("OAuth2 setting receive from NRF:", oauth2)
		}
	}
	if applyOAuth2 {
		udrContext.OAuth2Required = oauth2
		if oauth2 && udrContext.NrfCertPem == "" {
			logger.CfgLog.Error("OAuth2 enable but no nrfCertPem provided in config.")
		}
	} else if oauth2 != udrContext.OAuth2Required {
		logger.ConsumerLog.Warnf("NRF OAuth2 setting changed to %v, restart UDR to apply it", oauth2)
	}
}

// SendUpdateNFInstance sends an NFUpdate PATCH to the NRF, honoring ctx. The
// raw err comes back alongside any ProblemDetails so callers can read its
// GenericOpenAPIError status.
func (ns *NrfService) SendUpdateNFInstance(ctx context.Context, patchItems []models.PatchItem) (
	nf models.Nrf_NFMgmt_NFProfile, problemDetails *models.ProblemDetails, err error,
) {
	udrContext := udr_context.GetSelf()
	tokCtx, pd, err := udrContext.GetTokenCtx(
		models.Nrf_NFMgmt_ServiceName_NNRF_NFM,
		models.Nrf_NFMgmt_NFType_NRF,
	)
	if err != nil {
		return nf, pd, err
	}
	// GetTokenCtx takes no parent, so the token request stays uncancelable;
	// transplanting the token lets at least the PATCH honor ctx.
	if tok := tokCtx.Value(openapi.ContextOAuth2); tok != nil {
		ctx = context.WithValue(ctx, openapi.ContextOAuth2, tok)
	}

	client := ns.getNFManagementClient(udrContext.NrfUri)
	if client == nil {
		return nf, nil, fmt.Errorf("SendUpdateNFInstance: nrf not found")
	}

	request := &NFManagement.UpdateNFInstanceRequest{
		NfInstanceID: &udrContext.NfId,
		RequestBody:  patchItems,
	}

	res, err := client.NFInstanceIDDocumentApi.UpdateNFInstance(ctx, request)
	if err != nil {
		var apiErr openapi.GenericOpenAPIError
		if errors.As(err, &apiErr) {
			if updateErr, okModel := apiErr.Model().(NFManagement.UpdateNFInstanceError); okModel {
				return nf, updateErr.ProblemDetails, err
			}
		}
		return nf, nil, err
	}
	if res == nil {
		return nf, nil, fmt.Errorf("empty NFUpdate response")
	}
	if res.Nrf_NFMgmt_NFProfile != nil {
		nf = *res.Nrf_NFMgmt_NFProfile
	}
	return nf, nil, nil
}

func (ns *NrfService) SendDeregisterNFInstance() (err error) {
	logger.ConsumerLog.Infof("Send Deregister NFInstance")

	ctx, pd, err := udr_context.GetSelf().GetTokenCtx(models.Nrf_NFMgmt_ServiceName_NNRF_NFM, models.Nrf_NFMgmt_NFType_NRF)
	if err != nil {
		logger.ConsumerLog.Errorf("Get token context failed: problem details: %+v", pd)
		return err
	}

	udrSelf := udr_context.GetSelf()
	// Set client and set url
	configuration := NFManagement.NewConfiguration()
	configuration.SetBasePath(udrSelf.NrfUri)
	configuration.SetMetrics(sbi_metrics.SbiMetricHook)
	client := ns.getNFManagementClient(udrSelf.NrfUri)

	deregisterReq := &NFManagement.DeregisterNFInstanceRequest{
		NfInstanceID: &udrSelf.NfId,
	}
	_, deregisterErr := client.NFInstanceIDDocumentApi.DeregisterNFInstance(ctx, deregisterReq)
	if deregisterErr != nil {
		return deregisterErr
	}
	return nil
}

func (ns *NrfService) SendSearchNFInstances(nrfUri string,
	param NFDiscovery.SearchNFInstancesRequest,
) (*NFDiscovery.SearchNFInstancesResponse, error) {
	// Set client and set url
	configuration := NFDiscovery.NewConfiguration()
	configuration.SetBasePath(nrfUri)
	configuration.SetMetrics(sbi_metrics.SbiMetricHook)
	client := NFDiscovery.NewAPIClient(configuration)

	ctx, _, err := udr_context.GetSelf().GetTokenCtx(models.Nrf_NFMgmt_ServiceName_NNRF_DISC, models.Nrf_NFMgmt_NFType_NRF)
	if err != nil {
		return nil, err
	}

	result, err := client.NFInstancesStoreApi.SearchNFInstances(ctx, &param)

	return result, err
}
