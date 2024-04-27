package processor

import (
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/pkg/factory"
)

type Udr interface {
	Config() *factory.Config
	Context() *udr_context.UDRContext
}

type Processor struct {
	Udr
}

func NewProcessor(udr Udr) *Processor {
	return &Processor{
		Udr: udr,
	}
}