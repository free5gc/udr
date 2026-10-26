package processor

import (
	"reflect"
	"testing"

	"github.com/free5gc/openapi/models"
)

func TestNormalizeSMPolicySnssaiData(t *testing.T) {
	for _, tc := range []struct {
		name, oldKey string
	}{
		{"standard", "222-00007b"},
		{"legacy-decimal", "22200007b"},
		{"legacy-hex", "de00007b"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snssai := &models.Snssai{Sst: 222, Sd: "00007B"}
			entry := models.Udr_DR_SmPolicySnssaiData{
				Snssai: snssai,
				SmPolicyDnnData: map[string]models.Udr_DR_SmPolicyDnnData{
					"default": {Dnn: "default", GbrDl: "100 Mbps", GbrUl: "50 Mbps"},
				},
				UeSliceMbr: &models.SliceMbr{Downlink: "2 Gbps", Uplink: "1 Gbps"},
			}
			input := map[string]models.Udr_DR_SmPolicySnssaiData{tc.oldKey: entry}
			got, err := normalizeSMPolicySnssaiData(input)
			if err != nil {
				t.Fatal(err)
			}
			value, ok := got["222-00007b"]
			if !ok || len(got) != 1 {
				t.Fatalf("unexpected keys: %v", got)
			}
			if !reflect.DeepEqual(value.SmPolicyDnnData, entry.SmPolicyDnnData) ||
				!reflect.DeepEqual(value.UeSliceMbr, entry.UeSliceMbr) {
				t.Fatal("policy data changed")
			}
			if snssai.Sd != "00007B" || input[tc.oldKey].Snssai != snssai {
				t.Fatal("input mutated")
			}
			twice, err := normalizeSMPolicySnssaiData(got)
			if err != nil || !reflect.DeepEqual(twice, got) {
				t.Fatal("normalization is not idempotent")
			}
		})
	}
}

func TestNormalizeSMPolicySnssaiDataRejectsInvalidEntries(t *testing.T) {
	for _, snssai := range []*models.Snssai{nil, {Sst: 256}} {
		_, err := normalizeSMPolicySnssaiData(map[string]models.Udr_DR_SmPolicySnssaiData{"old": {Snssai: snssai}})
		if err == nil {
			t.Errorf("accepted invalid snssai: %+v", snssai)
		}
	}
	snssai := &models.Snssai{Sst: 222, Sd: "00007b"}
	if _, err := normalizeSMPolicySnssaiData(map[string]models.Udr_DR_SmPolicySnssaiData{
		"1-00007b": {Snssai: snssai},
	}); err == nil {
		t.Fatal("accepted contradictory key")
	}
	entries := map[string]models.Udr_DR_SmPolicySnssaiData{
		"de00007b":  {Snssai: snssai},
		"22200007b": {Snssai: snssai, UeSliceMbr: &models.SliceMbr{Downlink: "1 Gbps"}},
	}
	if _, err := normalizeSMPolicySnssaiData(entries); err == nil {
		t.Fatal("accepted conflicting aliases")
	}
	entries["22200007b"] = entries["de00007b"]
	got, err := normalizeSMPolicySnssaiData(entries)
	if err != nil || len(got) != 1 {
		t.Fatal("identical aliases should collapse")
	}
}
