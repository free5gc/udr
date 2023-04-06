package context

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/pkg/factory"
)

var udrContext = UDRContext{}

type subsId = string

type UDRServiceType int

const (
	NUDR_DR UDRServiceType = iota
)

func Init() {
	GetSelf().Name = "udr"
	GetSelf().EeSubscriptionIDGenerator = 1
	GetSelf().SdmSubscriptionIDGenerator = 1
	GetSelf().SubscriptionDataSubscriptionIDGenerator = 1
	GetSelf().PolicyDataSubscriptionIDGenerator = 1
	GetSelf().SubscriptionDataSubscriptions = make(map[subsId]*models.SubscriptionDataSubscriptions)
	GetSelf().PolicyDataSubscriptions = make(map[subsId]*models.PolicyDataSubscription)
	GetSelf().InfluenceDataSubscriptionIDGenerator = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
}

type UDRContext struct {
	Name                                    string
	UriScheme                               models.UriScheme
	BindingIPv4                             string
	SBIPort                                 int
	RegisterIPv4                            string // IP register to NRF
	HttpIPv6Address                         string
	NfId                                    string
	NrfUri                                  string
	EeSubscriptionIDGenerator               int
	SdmSubscriptionIDGenerator              int
	SubscriptionDataSubscriptionIDGenerator int
	PolicyDataSubscriptionIDGenerator       int
	InfluenceDataSubscriptionIDGenerator    *rand.Rand
	UESubsCollection                        sync.Map // map[ueId]*UESubsData
	UEGroupCollection                       sync.Map // map[ueGroupId]*UEGroupSubsData
	SubscriptionDataSubscriptions           map[subsId]*models.SubscriptionDataSubscriptions
	PolicyDataSubscriptions                 map[subsId]*models.PolicyDataSubscription
	InfluenceDataSubscriptions              sync.Map
	appDataInfluDataSubscriptionIdGenerator uint64
	mtx                                     sync.RWMutex
}

type UESubsData struct {
	EeSubscriptionCollection map[subsId]*EeSubscriptionCollection
	SdmSubscriptions         map[subsId]*models.SdmSubscription
}

type UEGroupSubsData struct {
	EeSubscriptions map[subsId]*models.EeSubscription
}

type EeSubscriptionCollection struct {
	EeSubscriptions      *models.EeSubscription
	AmfSubscriptionInfos []models.AmfSubscriptionInfo
}

// Reset UDR Context
func (context *UDRContext) Reset() {
	context.UESubsCollection.Range(func(key, value interface{}) bool {
		context.UESubsCollection.Delete(key)
		return true
	})
	context.UEGroupCollection.Range(func(key, value interface{}) bool {
		context.UEGroupCollection.Delete(key)
		return true
	})
	for key := range context.SubscriptionDataSubscriptions {
		delete(context.SubscriptionDataSubscriptions, key)
	}
	for key := range context.PolicyDataSubscriptions {
		delete(context.PolicyDataSubscriptions, key)
	}
	context.InfluenceDataSubscriptions.Range(func(key, value interface{}) bool {
		context.InfluenceDataSubscriptions.Delete(key)
		return true
	})
	context.EeSubscriptionIDGenerator = 1
	context.SdmSubscriptionIDGenerator = 1
	context.SubscriptionDataSubscriptionIDGenerator = 1
	context.PolicyDataSubscriptionIDGenerator = 1
	context.InfluenceDataSubscriptionIDGenerator = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	context.UriScheme = models.UriScheme_HTTPS
	context.Name = "udr"
}

func InitUdrContext(context *UDRContext) {
	config := factory.UdrConfig
	logger.UtilLog.Infof("udrconfig Info: Version[%s] Description[%s]", config.Info.Version, config.Info.Description)
	configuration := config.Configuration
	context.NfId = uuid.New().String()
	context.RegisterIPv4 = factory.UDR_DEFAULT_IPV4 // default localhost
	context.SBIPort = factory.UDR_DEFAULT_PORT_INT  // default port
	if sbi := configuration.Sbi; sbi != nil {
		context.UriScheme = models.UriScheme(sbi.Scheme)
		if sbi.RegisterIPv4 != "" {
			context.RegisterIPv4 = sbi.RegisterIPv4
		}
		if sbi.Port != 0 {
			context.SBIPort = sbi.Port
		}

		context.BindingIPv4 = os.Getenv(sbi.BindingIPv4)
		if context.BindingIPv4 != "" {
			logger.UtilLog.Info("Parsing ServerIPv4 address from ENV Variable.")
		} else {
			context.BindingIPv4 = sbi.BindingIPv4
			if context.BindingIPv4 == "" {
				logger.UtilLog.Warn("Error parsing ServerIPv4 address as string. Using the 0.0.0.0 address as default.")
				context.BindingIPv4 = "0.0.0.0"
			}
		}
	}
	if configuration.NrfUri != "" {
		context.NrfUri = configuration.NrfUri
	} else {
		logger.UtilLog.Warn("NRF Uri is empty! Using localhost as NRF IPv4 address.")
		context.NrfUri = fmt.Sprintf("%s://%s:%d", context.UriScheme, "127.0.0.1", 29510)
	}
}

func (context *UDRContext) GetIPv4Uri() string {
	return fmt.Sprintf("%s://%s:%d", context.UriScheme, context.RegisterIPv4, context.SBIPort)
}

func (context *UDRContext) GetIPv4GroupUri(udrServiceType UDRServiceType) string {
	var serviceUri string

	switch udrServiceType {
	case NUDR_DR:
		serviceUri = factory.UdrDrResUriPrefix
	default:
		serviceUri = ""
	}

	return fmt.Sprintf("%s://%s:%d%s", context.UriScheme, context.RegisterIPv4, context.SBIPort, serviceUri)
}

// Create new UDR context
func GetSelf() *UDRContext {
	return &udrContext
}

func (context *UDRContext) NewAppDataInfluDataSubscriptionID() uint64 {
	context.mtx.Lock()
	defer context.mtx.Unlock()
	context.appDataInfluDataSubscriptionIdGenerator++
	return context.appDataInfluDataSubscriptionIdGenerator
}

func NewInfluenceDataSubscriptionId() string {
	if GetSelf().InfluenceDataSubscriptionIDGenerator == nil {
		GetSelf().InfluenceDataSubscriptionIDGenerator = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	}
	return fmt.Sprintf("%08x", GetSelf().InfluenceDataSubscriptionIDGenerator.Uint32())
}
