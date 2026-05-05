package processor

import (
	"testing"

	"github.com/free5gc/openapi/models"
)

func TestCallbackServiceInfo(t *testing.T) {
	tests := []struct {
		name        string
		callbackURI string
		wantService models.ServiceName
		wantTarget  models.NrfNfManagementNfType
		wantOK      bool
	}{
		{
			name:        "pcf callback",
			callbackURI: "http://pcf.free5gc.org:8000/npcf-callback/v1/nudr-notify/policy-data/imsi-001010000000001",
			wantService: models.ServiceName("npcf-callback"),
			wantTarget:  models.NrfNfManagementNfType_PCF,
			wantOK:      true,
		},
		{
			name:        "nef callback",
			callbackURI: "http://nef.free5gc.org:8000/nnef-callback/v1/notification/smf",
			wantService: models.ServiceName("nnef-callback"),
			wantTarget:  models.NrfNfManagementNfType_NEF,
			wantOK:      true,
		},
		{
			name:        "smf callback without version segment",
			callbackURI: "http://smf.free5gc.org:8000/nsmf-callback/sm-policies/imsi-001010000000001-10",
			wantService: models.ServiceName("nsmf-callback"),
			wantTarget:  models.NrfNfManagementNfType_SMF,
			wantOK:      true,
		},
		{
			name:        "udm sdm subscription-data create callback",
			callbackURI: "http://udm.free5gc.org:8000/subscription-data/imsi-001010000000001/00101/sdm-subscriptions",
			wantService: models.ServiceName_NUDM_SDM,
			wantTarget:  models.NrfNfManagementNfType_UDM,
			wantOK:      true,
		},
		{
			name:        "udm sdm context-data update callback",
			callbackURI: "http://udm.free5gc.org:8000/subscription-data/imsi-001010000000001/context-data/sdm-subscriptions/subs-1",
			wantService: models.ServiceName_NUDM_SDM,
			wantTarget:  models.NrfNfManagementNfType_UDM,
			wantOK:      true,
		},
		{
			name:        "unknown path",
			callbackURI: "http://example.com/custom/notify",
			wantOK:      false,
		},
		{
			name:        "invalid uri",
			callbackURI: "://bad uri",
			wantOK:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotService, gotTarget, gotOK := callbackServiceInfo(tc.callbackURI)
			if gotOK != tc.wantOK {
				t.Fatalf("callbackServiceInfo(%q) ok = %v, want %v", tc.callbackURI, gotOK, tc.wantOK)
			}
			if gotService != tc.wantService {
				t.Fatalf("callbackServiceInfo(%q) service = %q, want %q", tc.callbackURI, gotService, tc.wantService)
			}
			if gotTarget != tc.wantTarget {
				t.Fatalf("callbackServiceInfo(%q) target = %q, want %q", tc.callbackURI, gotTarget, tc.wantTarget)
			}
		})
	}
}
