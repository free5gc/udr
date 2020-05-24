package consumer_test

import (
	"flag"
	"free5gc/lib/MongoDBLibrary"
	"free5gc/lib/openapi/Nnrf_NFDiscovery"
	"free5gc/lib/openapi/models"
	nrf_service "free5gc/src/nrf/service"
	"free5gc/src/udr/consumer"
	"free5gc/src/udr/factory"
	udr_service "free5gc/src/udr/service"
	"testing"
	"time"

	"github.com/antihax/optional"
	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
)

var flags flag.FlagSet
var c = cli.NewContext(nil, &flags, nil)

func nrfInit() {
	nrf := &nrf_service.NRF{}
	nrf.Initialize(c)
	go nrf.Start()
	time.Sleep(100 * time.Millisecond)
}
func TestRegisterNFInstance(t *testing.T) {
	// init NRF
	nrfInit()
	// Clear DB
	MongoDBLibrary.RestfulAPIDeleteMany("NfProfile", bson.M{})
	time.Sleep(50 * time.Millisecond)
	// Init UDR and register to NRF
	udr := udr_service.UDR{}
	udr.Initialize(c)
	go udr.Start()
	time.Sleep(50 * time.Millisecond)
	// Send NF Discovery to discover UDR
	param := Nnrf_NFDiscovery.SearchNFInstancesParamOpts{
		ServiceNames: optional.NewInterface([]models.ServiceName{models.ServiceName_NUDR_DR}),
	}
	result, err := consumer.SendSearchNFInstances(factory.UdrConfig.Configuration.NrfUri, models.NfType_UDR, models.NfType_UDR, param)
	if err != nil {
		t.Error(err.Error())
	} else if result.NfInstances == nil {
		t.Error("NF Instances is nil")
	}
}
