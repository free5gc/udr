package processor

import (
	"encoding/json"
	"github.com/free5gc/util/mongoapi"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi"
	"github.com/free5gc/udr/internal/util"
	"github.com/free5gc/udr/internal/logger"
)

const (
	APPDATA_INFLUDATA_DB_COLLECTION_NAME       = "applicationData.influenceData"
	APPDATA_INFLUDATA_SUBSC_DB_COLLECTION_NAME = "applicationData.influenceData.subsToNotify"
	APPDATA_PFD_DB_COLLECTION_NAME             = "applicationData.pfds"
)


func patchDataToDBAndNotify(collName string, ueId string, patchItem []models.PatchItem, filter bson.M) error {
	var err error
	origValue, err := mongoapi.RestfulAPIGetOne(collName, filter)
	if err != nil {
		return err
	}

	patchJSON, err := json.Marshal(patchItem)
	if err != nil {
		return err
	}

	if err = mongoapi.RestfulAPIJSONPatch(collName, filter, patchJSON); err != nil {
		return err
	}

	newValue, err := mongoapi.RestfulAPIGetOne(collName, filter)
	if err != nil {
		return err
	}
	PreHandleOnDataChangeNotify(ueId, CurrentResourceUri, patchItem, origValue, newValue)
	return nil
}

func getDataFromDB(collName string, filter bson.M) (map[string]interface{}, *models.ProblemDetails) {
	data, err := mongoapi.RestfulAPIGetOne(collName, filter)
	if err != nil {
		return nil, openapi.ProblemDetailsSystemFailure(err.Error())
	}
	if data == nil {
		return nil, util.ProblemDetailsNotFound("DATA_NOT_FOUND")
	}
	return data, nil
}

func getDataFromDBWithArg(collName string, filter bson.M, strength int) (
	map[string]interface{}, *models.ProblemDetails,
) {
	data, err := mongoapi.RestfulAPIGetOne(collName, filter, strength)
	if err != nil {
		return nil, openapi.ProblemDetailsSystemFailure(err.Error())
	}
	if data == nil {
		logger.ConsumerLog.Errorln("filter: ", filter)
		return nil, util.ProblemDetailsNotFound("DATA_NOT_FOUND")
	}

	return data, nil
}

func deleteDataFromDB(collName string, filter bson.M) {
	if err := mongoapi.RestfulAPIDeleteOne(collName, filter); err != nil {
		logger.DataRepoLog.Errorf("deleteDataFromDB: %+v", err)
	}
}