package consumer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/free5gc/openapi/models"
	NFDiscovery "github.com/free5gc/openapi/nrf/NFDisc"
	NFManagement "github.com/free5gc/openapi/nrf/NFMgmt"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	sbi_metrics "github.com/free5gc/util/metrics/sbi"
)

type NrfService struct {
	nfMngmntMu sync.RWMutex

	nfMngmntClients map[string]*NFManagement.APIClient
}

func (ns *NrfService) getNFManagementClient(uri string) *NFManagement.APIClient {
	if uri == "" {
		return nil
	}
	ns.nfMngmntMu.RLock()
	client, ok := ns.nfMngmntClients[uri]
	if ok {
		ns.nfMngmntMu.RUnlock()
		return client
	}

	configuration := NFManagement.NewConfiguration()
	configuration.SetBasePath(uri)
	configuration.SetMetrics(sbi_metrics.SbiMetricHook)
	client = NFManagement.NewAPIClient(configuration)

	ns.nfMngmntMu.RUnlock()
	ns.nfMngmntMu.Lock()
	defer ns.nfMngmntMu.Unlock()
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
	for serviceName, nfService := range context.NfService {
		if nfService.ServiceName == "" {
			nfService.ServiceName = serviceName
		}
		allowedNfTypes, known := udr_context.AllowedNfTypesForService(nfService.ServiceName)
		if !known {
			return profile, fmt.Errorf("no AllowedNfTypes policy for service %q", nfService.ServiceName)
		}
		nfService.AllowedNfTypes = allowedNfTypes
		services = append(services, nfService)
	}
	if len(services) > 0 {
		profile.NfServices = services
	}

	return profile, nil
}

func (ns *NrfService) SendRegisterNFInstance(ctx context.Context, nrfUri string) (
	resourceNrfUri string, retrieveNfInstanceId string, err error,
) {
	// Set client and set url
	profile, err := ns.buildNFProfile(udr_context.GetSelf())
	if err != nil {
		return "", "", fmt.Errorf("failed to build nrf profile %s", err.Error())
	}

	configuration := NFManagement.NewConfiguration()
	configuration.SetBasePath(nrfUri)
	configuration.SetMetrics(sbi_metrics.SbiMetricHook)
	client := ns.getNFManagementClient(nrfUri)

	finish := false

	for !finish {
		select {
		case <-ctx.Done():
			return "", "", fmt.Errorf("context done")
		default:
			registerReq := &NFManagement.RegisterNFInstanceRequest{
				NfInstanceID: &profile.NfInstanceId,
				RequestBody:  &profile,
			}
			rsp, registerErr := client.NFInstanceIDDocumentApi.RegisterNFInstance(ctx, registerReq)
			if registerErr != nil || rsp == nil {
				// TODO : add log
				logger.ConsumerLog.Errorf("UDR register to NRF Error[%v]", registerErr)
				time.Sleep(2 * time.Second)
				continue
			}

			resourceUri := rsp.Location
			resourceNrfUri, _, _ = strings.Cut(resourceUri, "/nnrf-nfm/")
			retrieveNfInstanceId = resourceUri[strings.LastIndex(resourceUri, "/")+1:]

			oauth2 := false

			if rsp.Nrf_NFMgmt_NFProfile != nil {
				customInfo, ok := rsp.Nrf_NFMgmt_NFProfile.CustomInfo.(map[string]interface{})
				v, isBool := customInfo["oauth2"].(bool)
				if ok && isBool {
					oauth2 = v
					logger.MainLog.Infoln("OAuth2 setting receive from NRF:", oauth2)
				}
			}
			if oauthErr := udr_context.GetSelf().SetOAuth2Required(oauth2); oauthErr != nil {
				return "", "", oauthErr
			}
			finish = true
		}
	}
	return resourceNrfUri, retrieveNfInstanceId, nil
}

func (ns *NrfService) SendDeregisterNFInstance() (err error) {
	logger.ConsumerLog.Infof("Send Deregister NFInstance")

	ctx, pd, err := udr_context.GetSelf().GetTokenCtxForNRF(models.Nrf_NFMgmt_ServiceName_NNRF_NFM)
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

	ctx, _, err := udr_context.GetSelf().GetTokenCtxForNRF(models.Nrf_NFMgmt_ServiceName_NNRF_DISC)
	if err != nil {
		return nil, err
	}

	result, err := client.NFInstancesStoreApi.SearchNFInstances(ctx, &param)

	return result, err
}
