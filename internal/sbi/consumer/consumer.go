package consumer

import (
	"github.com/free5gc/udr/pkg/app"
	"github.com/free5gc/openapi/Nnrf_NFManagement"
)

type Consumer struct {
	app.UdrApp
	*NrfService
}

func NewConsumer(udr app.UdrApp) *Consumer {
	configuration := Nnrf_NFManagement.NewConfiguration()
	configuration.SetBasePath(udr.Context().NrfUri)
	nrfService := &NrfService{
		nrfNfMgmtClient: Nnrf_NFManagement.NewAPIClient(configuration),
	}

	return &Consumer{
		UdrApp: udr,
		NrfService: nrfService,
	}
}