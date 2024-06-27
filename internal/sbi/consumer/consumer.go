package consumer

import (
	"github.com/free5gc/openapi/Nnrf_NFManagement"
	"github.com/free5gc/udr/pkg/app"
)

type Consumer struct {
	app.App

	*NrfService
}

func NewConsumer(udr app.App) *Consumer {
	configuration := Nnrf_NFManagement.NewConfiguration()
	configuration.SetBasePath(udr.Context().NrfUri)
	nrfService := &NrfService{
		nfMngmntClients: make(map[string]*Nnrf_NFManagement.APIClient),
	}

	return &Consumer{
		App:        udr,
		NrfService: nrfService,
	}
}
