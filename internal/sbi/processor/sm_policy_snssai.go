package processor

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
)

// normalizeSMPolicySnssaiData accepts legacy map keys using the mandatory
// snssai field as identity, and emits TS 29.571 section 5.4.4.2 map keys.
// It does not mutate the input, or change unrelated internal hex identifiers.
func normalizeSMPolicySnssaiData(entries map[string]models.Udr_DR_SmPolicySnssaiData) (
	map[string]models.Udr_DR_SmPolicySnssaiData, error,
) {
	if entries == nil {
		return nil, nil
	}
	result := make(map[string]models.Udr_DR_SmPolicySnssaiData, len(entries))
	for oldKey, entry := range entries {
		if entry.Snssai == nil {
			return nil, fmt.Errorf("SM policy entry %q is missing snssai", oldKey)
		}
		snssai, err := openapi.NormalizeSnssai(*entry.Snssai)
		if err != nil {
			return nil, fmt.Errorf("SM policy entry %q: %w", oldKey, err)
		}
		key, err := openapi.SnssaiModelsToMapKey(snssai)
		if err != nil {
			return nil, err
		}
		// A hyphenated key has an unambiguous standard identity. Do not silently
		// repair a contradictory identity. Unseparated legacy keys are ambiguous.
		if strings.Contains(oldKey, "-") && !strings.EqualFold(oldKey, key) &&
			(snssai.Sd != "" || !strings.EqualFold(oldKey, key+"-ffffff")) {
			return nil, fmt.Errorf("SM policy key %q disagrees with snssai", oldKey)
		}
		entry.Snssai = &snssai
		if previous, ok := result[key]; ok && !reflect.DeepEqual(previous, entry) {
			return nil, fmt.Errorf("conflicting SM policy entries for %q", key)
		}
		result[key] = entry
	}
	return result, nil
}
