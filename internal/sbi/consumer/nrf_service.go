package consumer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/Nnrf_NFManagement"
	"github.com/free5gc/openapi/Nnrf_NFDiscovery"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/udr/pkg/factory"
)

type NrfService struct {
	nrfNfMgmtClient *Nnrf_NFManagement.APIClient
}

func (ns *NrfService) buildNFInstance(context *udr_context.UDRContext) (models.NfProfile, error) {
	config := factory.UdrConfig

	profile := models.NfProfile{
		NfInstanceId:  context.NfId,
		NfType:        models.NfType_UDR,
		NfStatus:      models.NfStatus_REGISTERED,
		Ipv4Addresses: []string{context.RegisterIPv4},
		UdrInfo: &models.UdrInfo{
			SupportedDataSets: []models.DataSetId{
				// models.DataSetId_APPLICATION,
				// models.DataSetId_EXPOSURE,
				// models.DataSetId_POLICY,
				models.DataSetId_SUBSCRIPTION,
			},
		},
	}

	version := config.Info.Version
	tmpVersion := strings.Split(version, ".")
	versionUri := "v" + tmpVersion[0]
	apiPrefix := fmt.Sprintf("%s://%s:%d", context.UriScheme, context.RegisterIPv4, context.SBIPort)
	profile.NfServices = &[]models.NfService{
		{
			ServiceInstanceId: "datarepository",
			ServiceName:       models.ServiceName_NUDR_DR,
			Versions: &[]models.NfServiceVersion{
				{
					ApiFullVersion:  version,
					ApiVersionInUri: versionUri,
				},
			},
			Scheme:          context.UriScheme,
			NfServiceStatus: models.NfServiceStatus_REGISTERED,
			ApiPrefix:       apiPrefix,
			IpEndPoints: &[]models.IpEndPoint{
				{
					Ipv4Address: context.RegisterIPv4,
					Transport:   models.TransportProtocol_TCP,
					Port:        int32(context.SBIPort),
				},
			},
		},
	}

	// TODO: finish the Udr Info
	return profile, nil
}


func (ns *NrfService) SendRegisterNFInstance(ctx context.Context, nrfUri string) (string, string, error) {
	
	// Set client and set url
	profile, err := ns.buildNFInstance(udr_context.GetSelf())
	if err != nil {
		return "", "", fmt.Errorf("failed to build nrf profile %s", err.Error())
	}

	configuration := Nnrf_NFManagement.NewConfiguration()
	configuration.SetBasePath(nrfUri)
	client := Nnrf_NFManagement.NewAPIClient(configuration)
	var resouceNrfUri string
	var retrieveNfInstanceId string

	finish := false

	for !finish{
		select {
		case <-ctx.Done():
			return "", "", fmt.Errorf("context done")
		default:
			nf, res, err := client.NFInstanceIDDocumentApi.RegisterNFInstance(ctx, profile.NfInstanceId, profile)
			if err != nil || res == nil {
				// TODO : add log
				fmt.Println(fmt.Errorf("UDR register to NRF Error[%s]", err.Error()))
				time.Sleep(2 * time.Second)
				continue
			}
			defer func() {
				if rspCloseErr := res.Body.Close(); rspCloseErr != nil {
					logger.ConsumerLog.Errorf("RegisterNFInstance response body cannot close: %+v", rspCloseErr)
				}
			}()

			status := res.StatusCode
			if status == http.StatusOK {
				// NFUpdate
				finish = true
			} else if status == http.StatusCreated {
				// NFRegister
				resourceUri := res.Header.Get("Location")
				resouceNrfUri = resourceUri[:strings.Index(resourceUri, "/nnrf-nfm/")]
				retrieveNfInstanceId = resourceUri[strings.LastIndex(resourceUri, "/")+1:]

				oauth2 := false
				if nf.CustomInfo != nil {
					v, ok := nf.CustomInfo["oauth2"].(bool)
					if ok {
						oauth2 = v
						logger.MainLog.Infoln("OAuth2 setting receive from NRF:", oauth2)
					}
				}
				udr_context.GetSelf().OAuth2Required = oauth2
				if oauth2 && udr_context.GetSelf().NrfCertPem == "" {
					logger.CfgLog.Error("OAuth2 enable but no nrfCertPem provided in config.")
				}
				finish = true
			} else {
				fmt.Println("handler returned wrong status code", status)
				fmt.Println("NRF return wrong status code", status)
			}
		}
	}
	return resouceNrfUri, retrieveNfInstanceId, err 
}

func (ns *NrfService) SendDeregisterNFInstance() (problemDetails *models.ProblemDetails, err error) {
	logger.ConsumerLog.Infof("Send Deregister NFInstance")

	ctx, pd, err := udr_context.GetSelf().GetTokenCtx(models.ServiceName_NNRF_NFM, models.NfType_NRF)
	if err != nil {
		return pd, err
	}

	udrSelf := udr_context.GetSelf()
	// Set client and set url
	configuration := Nnrf_NFManagement.NewConfiguration()
	configuration.SetBasePath(udrSelf.NrfUri)
	client := Nnrf_NFManagement.NewAPIClient(configuration)

	var res *http.Response

	res, err = client.NFInstanceIDDocumentApi.DeregisterNFInstance(ctx, udrSelf.NfId)
	if err == nil {
		return nil, err
	} else if res != nil {
		defer func() {
			if rspCloseErr := res.Body.Close(); rspCloseErr != nil {
				logger.ConsumerLog.Errorf("DeregisterNFInstance response body cannot close: %+v", rspCloseErr)
			}
		}()

		if res.Status != err.Error() {
			return nil, err
		}
		problem := err.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		problemDetails = &problem
	} else {
		err = openapi.ReportError("server no response")
	}
	return problemDetails, err
}

func (ns *NrfService)SendSearchNFInstances(nrfUri string, targetNfType, requestNfType models.NfType,
	param Nnrf_NFDiscovery.SearchNFInstancesParamOpts,
) (*models.SearchResult, error) {
	// Set client and set url
	configuration := Nnrf_NFDiscovery.NewConfiguration()
	configuration.SetBasePath(nrfUri)
	client := Nnrf_NFDiscovery.NewAPIClient(configuration)

	ctx, _, err := udr_context.GetSelf().GetTokenCtx(models.ServiceName_NNRF_DISC, models.NfType_NRF)
	if err != nil {
		return nil, err
	}

	var res *http.Response
	result, res, err := client.NFInstancesStoreApi.SearchNFInstances(ctx, targetNfType, requestNfType, &param)
	if res != nil && res.StatusCode == http.StatusTemporaryRedirect {
		err = fmt.Errorf("temporary tedirect for non NRF consumer")
	}
	defer func() {
		if rspCloseErr := res.Body.Close(); rspCloseErr != nil {
			logger.ConsumerLog.Errorf("SearchNFInstances response body cannot close: %+v", rspCloseErr)
		}
	}()

	return &result, err
}