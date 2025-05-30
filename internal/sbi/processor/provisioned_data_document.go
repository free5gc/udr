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
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/internal/util"
	"github.com/free5gc/util/mongoapi"
)

func (p *Processor) QueryProvisionedDataProcedure(c *gin.Context, ueId string, servingPlmnId string,
	provisionedDataSets models.ProvisionedDataSets,
) {
	var collName string
	var filter bson.M

	collName = "subscriptionData.provisionedData.amData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	accessAndMobilitySubscriptionData, pd := p.GetDataFromDB(collName, filter)
	if pd != nil && pd.Status == http.StatusInternalServerError {
		logger.DataRepoLog.Errorf(
			"QueryProvisionedDataProcedure get accessAndMobilitySubscriptionData err: %s", pd.Detail)
		c.JSON(int(pd.Status), pd)
		return
	}
	if accessAndMobilitySubscriptionData != nil {
		var tmp models.AccessAndMobilitySubscriptionData
		if err := mapstructure.Decode(accessAndMobilitySubscriptionData, &tmp); err != nil {
			logger.DataRepoLog.Errorf(
				"QueryProvisionedDataProcedure accessAndMobilitySubscriptionData decode err: %+v", err)
			c.JSON(http.StatusInternalServerError, openapi.ProblemDetailsSystemFailure(err.Error()))
			return
		}
		provisionedDataSets.AmData = &tmp
	}

	collName = "subscriptionData.provisionedData.smfSelectionSubscriptionData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	smfSelectionSubscriptionData, pd := p.GetDataFromDB(collName, filter)
	if pd != nil && pd.Status == http.StatusInternalServerError {
		logger.DataRepoLog.Errorf("QueryProvisionedDataProcedure get smfSelectionSubscriptionData err: %s", pd.Detail)
		c.JSON(int(pd.Status), pd)
		return
	}
	if smfSelectionSubscriptionData != nil {
		var tmp models.SmfSelectionSubscriptionData
		if err := mapstructure.Decode(smfSelectionSubscriptionData, &tmp); err != nil {
			logger.DataRepoLog.Errorf(
				"QueryProvisionedDataProcedure smfSelectionSubscriptionData decode err: %+v", err)
			c.JSON(http.StatusInternalServerError, openapi.ProblemDetailsSystemFailure(err.Error()))
		}
		provisionedDataSets.SmfSelData = &tmp
	}

	collName = "subscriptionData.provisionedData.smsData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	smsSubscriptionData, pd := p.GetDataFromDB(collName, filter)
	if pd != nil && pd.Status == http.StatusInternalServerError {
		logger.DataRepoLog.Errorf("QueryProvisionedDataProcedure get smsSubscriptionData err: %s", pd.Detail)
		c.JSON(int(pd.Status), pd)
		return
	}
	if smsSubscriptionData != nil {
		var tmp models.SmsSubscriptionData
		if err := mapstructure.Decode(smsSubscriptionData, &tmp); err != nil {
			logger.DataRepoLog.Errorf(
				"QueryProvisionedDataProcedure smsSubscriptionData decode err: %+v", err)
			c.JSON(http.StatusInternalServerError, openapi.ProblemDetailsSystemFailure(err.Error()))
			return
		}
		provisionedDataSets.SmsSubsData = &tmp
	}

	collName = "subscriptionData.provisionedData.smData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	sessionManagementSubscriptionDatas, err := mongoapi.
		RestfulAPIGetMany(collName, filter, mongoapi.COLLATION_STRENGTH_IGNORE_CASE)
	if err != nil {
		logger.DataRepoLog.Errorf("QueryProvisionedDataProcedure get sessionManagementSubscriptionDatas err: %+v", err)
		c.JSON(http.StatusInternalServerError, openapi.ProblemDetailsSystemFailure(err.Error()))
		return
	}
	if sessionManagementSubscriptionDatas != nil {
		var tmp []models.SessionManagementSubscriptionData
		if err := mapstructure.Decode(sessionManagementSubscriptionDatas, &tmp); err != nil {
			logger.DataRepoLog.Errorf(
				"QueryProvisionedDataProcedure sessionManagementSubscriptionDatas decode err: %+v", err)
			c.JSON(http.StatusInternalServerError, openapi.ProblemDetailsSystemFailure(err.Error()))
			return
		}
		for _, smData := range tmp {
			dnnConfigurations := smData.DnnConfigurations
			tmpDnnConfigurations := make(map[string]models.DnnConfiguration)
			for escapedDnn, dnnConf := range dnnConfigurations {
				dnn := util.UnescapeDnn(escapedDnn)
				tmpDnnConfigurations[dnn] = dnnConf
			}
			smData.DnnConfigurations = tmpDnnConfigurations
		}
		provisionedDataSets.SmData.IndividualSmSubsData = tmp
	}

	collName = "subscriptionData.provisionedData.traceData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	traceData, pd := p.GetDataFromDB(collName, filter)
	if pd != nil && pd.Status == http.StatusInternalServerError {
		logger.DataRepoLog.Errorf("QueryProvisionedDataProcedure get traceData err: %s", pd.Detail)
		c.JSON(int(pd.Status), pd)
		return
	}
	if traceData != nil {
		var tmp models.TraceData
		if err := mapstructure.Decode(traceData, &tmp); err != nil {
			logger.DataRepoLog.Errorf("QueryProvisionedDataProcedure traceData decode err: %+v", err)
			c.JSON(http.StatusInternalServerError, openapi.ProblemDetailsSystemFailure(err.Error()))
			return
		}
		provisionedDataSets.TraceData = &tmp
	}

	collName = "subscriptionData.provisionedData.smsMngData"
	filter = bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}
	smsManagementSubscriptionData, pd := p.GetDataFromDB(collName, filter)
	if pd != nil && pd.Status == http.StatusInternalServerError {
		logger.DataRepoLog.Errorf(
			"QueryProvisionedDataProcedure get smsManagementSubscriptionData err: %s", pd.Detail)
		c.JSON(int(pd.Status), pd)
		return
	}
	if smsManagementSubscriptionData != nil {
		var tmp models.SmsManagementSubscriptionData
		if err := mapstructure.Decode(smsManagementSubscriptionData, &tmp); err != nil {
			logger.DataRepoLog.Errorf(
				"QueryProvisionedDataProcedure smsManagementSubscriptionData decode err: %+v", err)
			c.JSON(http.StatusInternalServerError, openapi.ProblemDetailsSystemFailure(err.Error()))
			return
		}
		provisionedDataSets.SmsMngData = &tmp
	}

	if reflect.DeepEqual(provisionedDataSets, models.ProvisionedDataSets{}) {
		pd := util.ProblemDetailsNotFound("DATA_NOT_FOUND")
		c.JSON(int(pd.Status), pd)
		return
	}
	c.JSON(http.StatusOK, provisionedDataSets)
}
