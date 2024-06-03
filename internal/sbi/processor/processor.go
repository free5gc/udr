package processor

import (
	"github.com/free5gc/udr/internal/database"
	"github.com/free5gc/udr/pkg/app"
)

type Processor struct {
	app.UdrApp
	database.DbConnector
}

func NewProcessor(udr app.UdrApp, dbInplement database.DbConnector) *Processor {
	return &Processor{
		UdrApp:      udr,
		DbConnector: dbInplement,
	}
}
