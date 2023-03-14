package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/udr/internal/logger"
)

func MapToByte(data map[string]interface{}) []byte {
	ret, err := json.Marshal(data)
	if err != nil {
		logger.UtilLog.Error(err)
	}
	return ret
}

func MapArrayToByte(data []map[string]interface{}) []byte {
	ret, err := json.Marshal(data)
	if err != nil {
		logger.UtilLog.Error(err)
	}
	return ret
}

func PrimitiveAToByte(data []interface{}) []byte {
	ret, err := json.Marshal(data)
	if err != nil {
		logger.UtilLog.Error(err)
	}
	return ret
}

func ToBsonM(data interface{}) bson.M {
	tmp, err := json.Marshal(data)
	if err != nil {
		logger.UtilLog.Error(err)
	}
	putData := bson.M{}
	err = json.Unmarshal(tmp, &putData)
	if err != nil {
		logger.UtilLog.Error(err)
	}
	return putData
}

func SnssaiHexToModels(hexString string) (*models.Snssai, error) {
	sst, err := strconv.ParseInt(hexString[:2], 16, 32)
	if err != nil {
		return nil, err
	}
	sNssai := &models.Snssai{
		Sst: int32(sst),
		Sd:  hexString[2:],
	}
	return sNssai, nil
}

func SnssaiModelsToHex(snssai models.Snssai) string {
	sst := fmt.Sprintf("%02X", snssai.Sst)
	return sst + snssai.Sd
}

func SnssaiModelsToTS29571String(snssai models.Snssai) string {
	sst := fmt.Sprintf("%d", snssai.Sst)
	if snssai.Sd != "" {
		return sst + "-" + snssai.Sd
	}
	return sst
}

func SnssaiHexToTS29571String(hexString string) (string, error) {
	if snssai, err := SnssaiHexToModels(hexString); err != nil {
		return "", err
	} else {
		return SnssaiModelsToTS29571String(*snssai), nil
	}
}

func EscapeDnn(dnn string) string {
	return strings.ReplaceAll(dnn, ".", "_")
}

func UnescapeDnn(dnnKey string) string {
	return strings.ReplaceAll(dnnKey, "_", ".")
}
