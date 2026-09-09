package context

import (
	"slices"

	"github.com/free5gc/openapi/models"
)

var servicePolicies = map[models.Nrf_NFMgmt_ServiceName][]models.Nrf_NFMgmt_NFType{
	models.Nrf_NFMgmt_ServiceName_NUDR_DR: {
		models.Nrf_NFMgmt_NFType_UDM,
		models.Nrf_NFMgmt_NFType_PCF,
		models.Nrf_NFMgmt_NFType_NEF,
		models.Nrf_NFMgmt_NFType_NRF,
		models.Nrf_NFMgmt_NFType_HSS,
	},
}

func AllowedNfTypesForService(
	serviceName models.Nrf_NFMgmt_ServiceName,
) ([]models.Nrf_NFMgmt_NFType, bool) {
	allowed, known := servicePolicies[serviceName]
	return slices.Clone(allowed), known
}
