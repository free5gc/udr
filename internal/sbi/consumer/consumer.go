package consumer

import (
	NFManagement "github.com/free5gc/openapi/nrf/NFMgmt"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/pkg/app"
	"github.com/free5gc/util/nfheartbeat"
)

type Consumer struct {
	app.App

	*NrfService
}

func NewConsumer(udr app.App) (*Consumer, error) {
	configuration := NFManagement.NewConfiguration()
	configuration.SetBasePath(udr.Context().NrfUri)
	nrfService := &NrfService{
		nfMngmntClients: make(map[string]*NFManagement.APIClient),
	}

	c := &Consumer{
		App:        udr,
		NrfService: nrfService,
	}
	heartbeat, err := nfheartbeat.NewRunner(
		nrfRegistrar{nrfService},
		func() int32 { return udr.Config().GetNfHeartBeatTimer() },
		logger.ConsumerLog,
	)
	if err != nil {
		return nil, err
	}
	nrfService.heartbeat = heartbeat

	return c, nil
}
