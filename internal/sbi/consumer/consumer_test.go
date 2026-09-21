package consumer

import (
	"testing"

	"github.com/stretchr/testify/require"

	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/pkg/app"
	"github.com/free5gc/udr/pkg/factory"
)

var _ app.App = (*fakeApp)(nil)

// fakeApp is the app.App the consumer needs. UDR has no generated mock for it:
// the only mock in the tree covers the SBI server interface.
type fakeApp struct {
	cfg *factory.Config
}

func (a *fakeApp) SetLogEnable(enable bool) {}

func (a *fakeApp) SetLogLevel(level string) {}

func (a *fakeApp) SetReportCaller(reportCaller bool) {}

func (a *fakeApp) Start() {}

func (a *fakeApp) Terminate() {}

// NrfService reads the context global directly, so handing back anything else
// would let the two views of the NF profile drift apart.
func (a *fakeApp) Context() *udr_context.UDRContext { return udr_context.GetSelf() }

// The same pointer every call: the fallback-interval test mutates the config.
func (a *fakeApp) Config() *factory.Config { return a.cfg }

// newTestConsumer builds a Consumer over the UDR context global, pointed at nfId
// and nrfUri for the duration of the test.
func newTestConsumer(t *testing.T, nfId, nrfUri string) *Consumer {
	t.Helper()

	udrContext := udr_context.GetSelf()
	prevNfId, prevNrfUri, prevOAuth2 := udrContext.NfId, udrContext.NrfUri, udrContext.OAuth2Required
	t.Cleanup(func() {
		udrContext.NfId = prevNfId
		udrContext.NrfUri = prevNrfUri
		udrContext.OAuth2Required = prevOAuth2
	})
	udrContext.NfId = nfId
	udrContext.NrfUri = nrfUri
	udrContext.OAuth2Required = false

	testConsumer, err := NewConsumer(&fakeApp{
		cfg: &factory.Config{Configuration: &factory.Configuration{}},
	})
	require.NoError(t, err)

	return testConsumer
}
