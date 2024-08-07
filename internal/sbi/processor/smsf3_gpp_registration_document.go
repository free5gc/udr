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
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/internal/util"
	"github.com/free5gc/util/mongoapi"
)

func (p *Processor) CreateSmsfContext3gppProcedure(
	c *gin.Context, collName string, ueId string, SmsfRegistration models.SmsfRegistration,
) {
	putData := util.ToBsonM(SmsfRegistration)
	putData["ueId"] = ueId
	filter := bson.M{"ueId": ueId}

	_, err := mongoapi.RestfulAPIPutOne(collName, filter, putData)
	if err != nil {
		logger.DataRepoLog.Errorf("CreateSmsfContext3gppProcedure err: %+v", err)
	}

	c.Status(http.StatusNoContent)
}

func (p *Processor) DeleteSmsfContext3gppProcedure(c *gin.Context, collName string, ueId string) {
	filter := bson.M{"ueId": ueId}
	p.DeleteDataFromDB(collName, filter)
	c.Status(http.StatusNoContent)
}

func (p *Processor) QuerySmsfContext3gppProcedure(c *gin.Context, collName string, ueId string) {
	filter := bson.M{"ueId": ueId}
	data, pd := p.GetDataFromDB(collName, filter)
	if pd != nil {
		logger.DataRepoLog.Errorf("QuerySmsfContext3gppProcedure err: %s", pd.Detail)
		c.JSON(int(pd.Status), pd)
		return
	}
	c.JSON(http.StatusOK, data)
}
