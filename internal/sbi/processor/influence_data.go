/*
 * Nudr_DataRepository API OpenAPI file
 *
 * Unified Data Repository Service
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package processor

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/free5gc/openapi/models"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/util/mongoapi"
)

func (p *Processor) ApplicationDataInfluenceDataGetProcedure(c *gin.Context, collName string, filter []bson.M) (
	response *[]map[string]interface{},
) {
	influenceDataArray := make([]map[string]interface{}, 0)
	if len(filter) != 0 {
		var err error
		influenceDataArray, err = mongoapi.RestfulAPIGetMany(collName, bson.M{"$and": filter})
		if err != nil {
			logger.DataRepoLog.Errorf("ApplicationDataInfluenceDataGetProcedure err: %+v", err)
			return nil
		}
	}
	for _, influenceData := range influenceDataArray {
		groupUri := udr_context.GetSelf().GetIPv4GroupUri(udr_context.NUDR_DR)
		influenceData["resUri"] = fmt.Sprintf("%s/application-data/influenceData/%s",
			groupUri, influenceData["influenceId"].(string))
		delete(influenceData, "_id")
		delete(influenceData, "influenceId")
	}
	c.JSON(http.StatusOK, influenceDataArray)
	return
}

func (p *Processor) ParseSnssaisFromQueryParam(snssaiStr string) []models.Snssai {
	var snssais []models.Snssai
	err := json.Unmarshal([]byte(snssaiStr), &snssais)
	if err != nil {
		logger.DataRepoLog.Warnln("Unmarshal Error in snssaiStruct", err)
	}
	return snssais
}

func (p *Processor) BuildSnssaiMatchList(snssais []models.Snssai) (matchList []bson.M) {
	for _, v := range snssais {
		matchList = append(matchList, bson.M{"snssai.sst": v.Sst, "snssai.sd": v.Sd})
	}
	return
}
