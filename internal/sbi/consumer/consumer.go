package consumer

import (
	"github.com/free5gc/openapi/Nnrf_NFManagement"
	"github.com/free5gc/udr/pkg/app"
)

type Consumer struct {
	app.UdrApp

	*NrfService
}

func NewConsumer(udr app.UdrApp) *Consumer {
	configuration := Nnrf_NFManagement.NewConfiguration()
	configuration.SetBasePath(udr.Context().NrfUri)
	nrfService := &NrfService{
		nfMngmntClients: make(map[string]*Nnrf_NFManagement.APIClient),
	}

	return &Consumer{
		UdrApp:     udr,
		NrfService: nrfService,
	}
}
