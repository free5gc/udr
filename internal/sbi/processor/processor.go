package processor

import (
	"github.com/free5gc/udr/pkg/app"
)



type Processor struct {
	app.UdrApp
}

func NewProcessor(udr app.UdrApp) *Processor {
	return &Processor{
		UdrApp: udr,
	}
}