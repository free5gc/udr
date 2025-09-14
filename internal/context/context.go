package context

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi/oauth"
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
	udrContext.Name = "udr"
	udrContext.EeSubscriptionIDGenerator = 1
	udrContext.SdmSubscriptionIDGenerator = 1
	udrContext.SubscriptionDataSubscriptionIDGenerator = 1
	udrContext.PolicyDataSubscriptionIDGenerator = 1
	udrContext.SubscriptionDataSubscriptions = make(map[subsId]*models.SubscriptionDataSubscriptions)
	udrContext.PolicyDataSubscriptions = make(map[subsId]*models.PolicyDataSubscription)
	udrContext.InfluenceDataSubscriptionIDGenerator = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	serviceName := []models.ServiceName{
		models.ServiceName_NUDR_DR,
	}
	udrContext.NrfUri = udrContext.GetIPUriWithPort(29510)
	initUdrContext()

	config := factory.UdrConfig
	udrContext.NfService = initNfService(serviceName, config.Info.Version)
}

type UDRContext struct {
	Name                                    string
	UriScheme                               models.UriScheme
	BindingIP                               netip.Addr
	RegisterIP                              netip.Addr // IP register to NRF
	SBIPort                                 int
	NfService                               map[models.ServiceName]models.NrfNfManagementNfService
	HttpIPv6Address                         string
	NfId                                    string
	NrfUri                                  string
	NrfCertPem                              string
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
	OAuth2Required                          bool
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

type NFContext interface {
	AuthorizationCheck(token string, serviceName models.ServiceName) error
}

var _ NFContext = &UDRContext{}

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

func initUdrContext() {
	config := factory.UdrConfig
	logger.UtilLog.Infof("udrconfig Info: Version[%s] Description[%s]", config.Info.Version, config.Info.Description)
	configuration := config.Configuration
	udrContext.NfId = uuid.New().String()

	sbi := configuration.Sbi

	udrContext.SBIPort = sbi.Port

	udrContext.UriScheme = models.UriScheme(sbi.Scheme)

	if bindingIP := os.Getenv(sbi.BindingIP); bindingIP != "" {
		logger.UtilLog.Info("Parsing BindingIP address from ENV Variable.")
		sbi.BindingIP = bindingIP
	}
	if registerIP := os.Getenv(sbi.RegisterIP); registerIP != "" {
		logger.UtilLog.Info("Parsing RegisterIP address from ENV Variable.")
		sbi.RegisterIP = registerIP
	}
	udrContext.BindingIP = resolveIP(sbi.BindingIP)
	udrContext.RegisterIP = resolveIP(sbi.RegisterIP)

	if configuration.NrfUri != "" {
		udrContext.NrfUri = configuration.NrfUri
	} else {
		logger.UtilLog.Warn("NRF Uri is empty! Using localhost as NRF IPv4 address.")
		udrContext.NrfUri = fmt.Sprintf("%s://%s:%d", udrContext.UriScheme, "127.0.0.1", 29510)
	}
	udrContext.NrfCertPem = configuration.NrfCertPem
}

func resolveIP(ip string) netip.Addr {
	resolvedIPs, err := net.DefaultResolver.LookupNetIP(context.Background(), "ip", ip)
	if err != nil {
		logger.InitLog.Errorf("Lookup failed with %s: %+v", ip, err)
	}
	resolvedIP := resolvedIPs[0].Unmap()
	if resolvedIP := resolvedIP.String(); resolvedIP != ip {
		logger.UtilLog.Infof("Lookup revolved %s into %s", ip, resolvedIP)
	}
	return resolvedIP
}

func (c *UDRContext) GetIPUri() string {
	addr := c.RegisterIP
	port := c.SBIPort

	return fmt.Sprintf("%s://%s", c.UriScheme, netip.AddrPortFrom(addr, uint16(port)).String())
}

func (c *UDRContext) GetIpEndPoint() []models.IpEndPoint {
	if c.RegisterIP.Is6() {
		return []models.IpEndPoint{
			{
				Ipv6Address: c.RegisterIP.String(),
				Transport:   models.NrfNfManagementTransportProtocol_TCP,
				Port:        int32(c.SBIPort),
			},
		}
	} else if c.RegisterIP.Is4() {
		return []models.IpEndPoint{
			{
				Ipv4Address: c.RegisterIP.String(),
				Transport:   models.NrfNfManagementTransportProtocol_TCP,
				Port:        int32(c.SBIPort),
			},
		}
	}
	return nil
}

func initNfService(serviceName []models.ServiceName, version string) (
	nfService map[models.ServiceName]models.NrfNfManagementNfService,
) {
	versionUri := "v" + strings.Split(version, ".")[0]
	nfService = make(map[models.ServiceName]models.NrfNfManagementNfService)
	for idx, name := range serviceName {
		nfService[name] = models.NrfNfManagementNfService{
			ServiceInstanceId: strconv.Itoa(idx),
			ServiceName:       name,
			Versions: []models.NfServiceVersion{
				{
					ApiFullVersion:  version,
					ApiVersionInUri: versionUri,
				},
			},
			Scheme:          udrContext.UriScheme,
			NfServiceStatus: models.NfServiceStatus_REGISTERED,
			ApiPrefix:       udrContext.GetIPUri(),
			IpEndPoints:     udrContext.GetIpEndPoint(),
		}
	}

	return
}

func (context *UDRContext) GetIPUriWithPort(port int) string {
	addr := context.RegisterIP

	return fmt.Sprintf("%s://%s", context.UriScheme, netip.AddrPortFrom(addr, uint16(port)).String())
}

func (context *UDRContext) GetIPv4GroupUri(udrServiceType UDRServiceType) string {
	var serviceUri string

	switch udrServiceType {
	case NUDR_DR:
		serviceUri = factory.UdrDrResUriPrefix
	default:
		serviceUri = ""
	}

	return fmt.Sprintf("%s%s", context.GetIPUri(), serviceUri)
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

func (c *UDRContext) GetTokenCtx(serviceName models.ServiceName, targetNF models.NrfNfManagementNfType) (
	context.Context, *models.ProblemDetails, error,
) {
	if !c.OAuth2Required {
		return context.TODO(), nil, nil
	}
	return oauth.GetTokenCtx(models.NrfNfManagementNfType_UDR, targetNF,
		c.NfId, c.NrfUri, string(serviceName))
}

func (c *UDRContext) AuthorizationCheck(token string, serviceName models.ServiceName) error {
	if !c.OAuth2Required {
		logger.UtilLog.Debugf("UDRContext::AuthorizationCheck: OAuth2 not required\n")
		return nil
	}

	logger.UtilLog.Debugf("UDRContext::AuthorizationCheck: token[%s] serviceName[%s]\n", token, serviceName)
	return oauth.VerifyOAuth(token, string(serviceName), c.NrfCertPem)
}
