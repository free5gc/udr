package database

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/free5gc/openapi/models"
)

const (
	APPDATA_INFLUDATA_DB_COLLECTION_NAME       = "applicationData.influenceData"
	APPDATA_INFLUDATA_SUBSC_DB_COLLECTION_NAME = "applicationData.influenceData.subsToNotify"
	APPDATA_PFD_DB_COLLECTION_NAME             = "applicationData.pfds"
)

type DbConnector interface {
	PatchDataToDBAndNotify(collName string, ueId string, patchItem []models.PatchItem, filter bson.M) error
	GetDataFromDB(collName string, filter bson.M) (map[string]interface{}, *models.ProblemDetails)
	GetDataFromDBWithArg(collName string, filter bson.M, strength int) (map[string]interface{}, *models.ProblemDetails)
	DeleteDataFromDB(collName string, filter bson.M)
}
