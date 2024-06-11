package processor

import (
	"github.com/free5gc/udr/internal/database"
	"github.com/free5gc/udr/pkg/app"
	"github.com/free5gc/udr/pkg/factory"
)

type Processor struct {
	app.UdrApp
	database.DbConnector
}

func NewProcessor(udr app.UdrApp) *Processor {
	return &Processor{
		UdrApp:      udr,
		DbConnector: database.NewDbConnector(factory.UdrConfig.Configuration.DbConnectorType),
	}
}
