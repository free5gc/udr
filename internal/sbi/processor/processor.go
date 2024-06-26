package processor

import (
	"github.com/free5gc/udr/internal/database"
	"github.com/free5gc/udr/pkg/app"
	"github.com/free5gc/udr/pkg/factory"
)

type Processor struct {
	app.App
	database.DbConnector
}

func NewProcessor(udr app.App) *Processor {
	return &Processor{
		App:         udr,
		DbConnector: database.NewDbConnector(factory.UdrConfig.Configuration.DbConnectorType),
	}
}
