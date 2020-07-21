package handler

import (
	"encoding/json"
	"free5gc/lib/openapi/models"
	udr_context "free5gc/src/udr/context"
	"free5gc/src/udr/handler/message"
	"free5gc/src/udr/logger"
	"free5gc/src/udr/producer"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var HandlerLog *logrus.Entry

func init() {
	// init Pool
	HandlerLog = logger.HandlerLog
}

func Handle() {
	for {
		select {
		case msg, ok := <-message.UdrChannel:
			if ok {
				producer.CurrentResourceUri = udr_context.UDR_Self().GetIPv4Uri() + msg.HTTPRequest.URL.EscapedPath()
				switch msg.Event {
				case message.EventCreateAccessAndMobilityData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleCreateAccessAndMobilityData(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.AccessAndMobilityData))
				case message.EventDeleteAccessAndMobilityData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleDeleteAccessAndMobilityData(msg.ResponseChan, ueId)
				case message.EventQueryAccessAndMobilityData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQueryAccessAndMobilityData(msg.ResponseChan, ueId)
				case message.EventAmfContext3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleAmfContext3gpp(msg.ResponseChan, ueId, msg.HTTPRequest.Body.([]models.PatchItem))
				case message.EventCreateAmfContext3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleCreateAmfContext3gpp(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.Amf3GppAccessRegistration))
				case message.EventQueryAmfContext3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQueryAmfContext3gpp(msg.ResponseChan, ueId)
				case message.EventAmfContextNon3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleAmfContextNon3gpp(msg.ResponseChan, ueId, msg.HTTPRequest.Body.([]models.PatchItem))
				case message.EventCreateAmfContextNon3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleCreateAmfContextNon3gpp(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.AmfNon3GppAccessRegistration))
				case message.EventQueryAmfContextNon3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQueryAmfContextNon3gpp(msg.ResponseChan, ueId)
				case message.EventModifyAmfSubscriptionInfo:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleModifyAmfSubscriptionInfo(msg.ResponseChan, ueId, subsId, msg.HTTPRequest.Body.([]models.PatchItem))
				case message.EventModifyAuthentication:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleModifyAuthentication(msg.ResponseChan, ueId, msg.HTTPRequest.Body.([]models.PatchItem))
				case message.EventQueryAuthSubsData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQueryAuthSubsData(msg.ResponseChan, ueId)
				case message.EventCreateAuthenticationSoR:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleCreateAuthenticationSoR(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.SorData))
				case message.EventQueryAuthSoR:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQueryAuthSoR(msg.ResponseChan, ueId)
				case message.EventCreateAuthenticationStatus:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleCreateAuthenticationStatus(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.AuthEvent))
				case message.EventQueryAuthenticationStatus:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQueryAuthenticationStatus(msg.ResponseChan, ueId)
				case message.EventApplicationDataInfluenceDataGet:
					// TODO
					producer.HandleApplicationDataInfluenceDataGet(msg.ResponseChan)
				case message.EventApplicationDataInfluenceDataInfluenceIdDelete:
					// TODO
					influenceId := msg.HTTPRequest.Params["influenceId"]
					producer.HandleApplicationDataInfluenceDataInfluenceIdDelete(msg.ResponseChan, influenceId)
				case message.EventApplicationDataInfluenceDataInfluenceIdPatch:
					// TODO
					influenceId := msg.HTTPRequest.Params["influenceId"]
					producer.HandleApplicationDataInfluenceDataInfluenceIdPatch(msg.ResponseChan, influenceId, msg.HTTPRequest.Body.(models.TrafficInfluDataPatch))
				case message.EventApplicationDataInfluenceDataInfluenceIdPut:
					// TODO
					influenceId := msg.HTTPRequest.Params["influenceId"]
					producer.HandleApplicationDataInfluenceDataInfluenceIdPut(msg.ResponseChan, influenceId, msg.HTTPRequest.Body.(models.TrafficInfluData))
				case message.EventApplicationDataInfluenceDataSubsToNotifyGet:
					// TODO
					producer.HandleApplicationDataInfluenceDataSubsToNotifyGet(msg.ResponseChan)
				case message.EventApplicationDataInfluenceDataSubsToNotifyPost:
					// TODO
					producer.HandleApplicationDataInfluenceDataSubsToNotifyPost(msg.ResponseChan, msg.HTTPRequest.Body.(models.TrafficInfluSub))
				case message.EventApplicationDataInfluenceDataSubsToNotifySubscriptionIdDelete:
					// TODO
					subscriptionId := msg.HTTPRequest.Params["subscriptionId"]
					producer.HandleApplicationDataInfluenceDataSubsToNotifySubscriptionIdDelete(msg.ResponseChan, subscriptionId)
				case message.EventApplicationDataInfluenceDataSubsToNotifySubscriptionIdGet:
					// TODO
					subscriptionId := msg.HTTPRequest.Params["subscriptionId"]
					producer.HandleApplicationDataInfluenceDataSubsToNotifySubscriptionIdGet(msg.ResponseChan, subscriptionId)
				case message.EventApplicationDataInfluenceDataSubsToNotifySubscriptionIdPut:
					// TODO
					subscriptionId := msg.HTTPRequest.Params["subscriptionId"]
					producer.HandleApplicationDataInfluenceDataSubsToNotifySubscriptionIdPut(msg.ResponseChan, subscriptionId, msg.HTTPRequest.Body.(models.TrafficInfluSub))
				case message.EventApplicationDataPfdsAppIdDelete:
					// TODO
					appId := msg.HTTPRequest.Params["appId"]
					producer.HandleApplicationDataPfdsAppIdDelete(msg.ResponseChan, appId)
				case message.EventApplicationDataPfdsAppIdGet:
					// TODO
					appId := msg.HTTPRequest.Params["appId"]
					producer.HandleApplicationDataPfdsAppIdGet(msg.ResponseChan, appId)
				case message.EventApplicationDataPfdsAppIdPut:
					// TODO
					appId := msg.HTTPRequest.Params["appId"]
					producer.HandleApplicationDataPfdsAppIdPut(msg.ResponseChan, appId, msg.HTTPRequest.Body.(models.PfdDataForApp))
				case message.EventApplicationDataPfdsGet:
					// TODO
					appIdArray := msg.HTTPRequest.Query["appId"]
					producer.HandleApplicationDataPfdsGet(msg.ResponseChan, appIdArray)
				case message.EventExposureDataSubsToNotifyPost:
					// TODO
					producer.HandleExposureDataSubsToNotifyPost(msg.ResponseChan, msg.HTTPRequest.Body.(models.ExposureDataSubscription))
				case message.EventExposureDataSubsToNotifySubIdDelete:
					// TODO
					subId := msg.HTTPRequest.Params["subId"]
					producer.HandleExposureDataSubsToNotifySubIdDelete(msg.ResponseChan, subId)
				case message.EventExposureDataSubsToNotifySubIdPut:
					// TODO
					subId := msg.HTTPRequest.Params["subId"]
					producer.HandleExposureDataSubsToNotifySubIdPut(msg.ResponseChan, subId, msg.HTTPRequest.Body.(models.ExposureDataSubscription))
				case message.EventPolicyDataBdtDataBdtReferenceIdDelete:
					// TODO
					bdtReferenceId := msg.HTTPRequest.Params["bdtReferenceId"]
					producer.HandlePolicyDataBdtDataBdtReferenceIdDelete(msg.ResponseChan, bdtReferenceId)
				case message.EventPolicyDataBdtDataBdtReferenceIdGet:
					// TODO
					bdtReferenceId := msg.HTTPRequest.Params["bdtReferenceId"]
					producer.HandlePolicyDataBdtDataBdtReferenceIdGet(msg.ResponseChan, bdtReferenceId)
				case message.EventPolicyDataBdtDataBdtReferenceIdPut:
					// TODO
					bdtReferenceId := msg.HTTPRequest.Params["bdtReferenceId"]
					producer.HandlePolicyDataBdtDataBdtReferenceIdPut(msg.ResponseChan, bdtReferenceId, msg.HTTPRequest.Body.(models.BdtData))
				case message.EventPolicyDataBdtDataGet:
					// TODO
					producer.HandlePolicyDataBdtDataGet(msg.ResponseChan)
				case message.EventPolicyDataPlmnsPlmnIdUePolicySetGet:
					// TODO
					plmnId := msg.HTTPRequest.Params["plmnId"]
					producer.HandlePolicyDataPlmnsPlmnIdUePolicySetGet(msg.ResponseChan, plmnId)
				case message.EventPolicyDataSponsorConnectivityDataSponsorIdGet:
					// TODO
					sponsorId := msg.HTTPRequest.Params["sponsorId"]
					producer.HandlePolicyDataSponsorConnectivityDataSponsorIdGet(msg.ResponseChan, sponsorId)
				case message.EventPolicyDataSubsToNotifyPost:
					// TODO
					producer.HandlePolicyDataSubsToNotifyPost(msg.ResponseChan, msg.HTTPRequest.Body.(models.PolicyDataSubscription))
				case message.EventPolicyDataSubsToNotifySubsIdDelete:
					// TODO
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandlePolicyDataSubsToNotifySubsIdDelete(msg.ResponseChan, subsId)
				case message.EventPolicyDataSubsToNotifySubsIdPut:
					// TODO
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandlePolicyDataSubsToNotifySubsIdPut(msg.ResponseChan, subsId, msg.HTTPRequest.Body.(models.PolicyDataSubscription))
				case message.EventPolicyDataUesUeIdAmDataGet:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandlePolicyDataUesUeIdAmDataGet(msg.ResponseChan, ueId)
				case message.EventPolicyDataUesUeIdOperatorSpecificDataGet:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandlePolicyDataUesUeIdOperatorSpecificDataGet(msg.ResponseChan, ueId)
				case message.EventPolicyDataUesUeIdOperatorSpecificDataPatch:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandlePolicyDataUesUeIdOperatorSpecificDataPatch(msg.ResponseChan, ueId, msg.HTTPRequest.Body.([]models.PatchItem))
				case message.EventPolicyDataUesUeIdOperatorSpecificDataPut:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandlePolicyDataUesUeIdOperatorSpecificDataPut(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(map[string]models.OperatorSpecificDataContainer))
				case message.EventPolicyDataUesUeIdSmDataGet:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]

					sNssai := models.Snssai{}
					sNssaiQuery := msg.HTTPRequest.Query.Get("snssai")
					err := json.Unmarshal([]byte(sNssaiQuery), &sNssai)
					if err != nil {
						HandlerLog.Warnln(err)
					}

					dnn := msg.HTTPRequest.Query.Get("dnn")

					producer.HandlePolicyDataUesUeIdSmDataGet(msg.ResponseChan, ueId, sNssai, dnn)
				case message.EventPolicyDataUesUeIdSmDataPatch:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandlePolicyDataUesUeIdSmDataPatch(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(map[string]models.UsageMonData))
				case message.EventPolicyDataUesUeIdSmDataUsageMonIdDelete:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					usageMonId := msg.HTTPRequest.Params["usageMonId"]
					producer.HandlePolicyDataUesUeIdSmDataUsageMonIdDelete(msg.ResponseChan, ueId, usageMonId)
				case message.EventPolicyDataUesUeIdSmDataUsageMonIdGet:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					usageMonId := msg.HTTPRequest.Params["usageMonId"]
					producer.HandlePolicyDataUesUeIdSmDataUsageMonIdGet(msg.ResponseChan, ueId, usageMonId)
				case message.EventPolicyDataUesUeIdSmDataUsageMonIdPut:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					usageMonId := msg.HTTPRequest.Params["usageMonId"]
					producer.HandlePolicyDataUesUeIdSmDataUsageMonIdPut(msg.ResponseChan, ueId, usageMonId, msg.HTTPRequest.Body.(models.UsageMonData))
				case message.EventPolicyDataUesUeIdUePolicySetGet:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandlePolicyDataUesUeIdUePolicySetGet(msg.ResponseChan, ueId)
				case message.EventPolicyDataUesUeIdUePolicySetPatch:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandlePolicyDataUesUeIdUePolicySetPatch(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.UePolicySet))
				case message.EventPolicyDataUesUeIdUePolicySetPut:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandlePolicyDataUesUeIdUePolicySetPut(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.UePolicySet))
				case message.EventCreateAMFSubscriptions:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleCreateAMFSubscriptions(msg.ResponseChan, ueId, subsId, msg.HTTPRequest.Body.([]models.AmfSubscriptionInfo))
				case message.EventRemoveAmfSubscriptionsInfo:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleRemoveAmfSubscriptionsInfo(msg.ResponseChan, ueId, subsId)
				case message.EventQueryEEData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQueryEEData(msg.ResponseChan, ueId)
				case message.EventRemoveEeGroupSubscriptions:
					// TODO
					ueGroupId := msg.HTTPRequest.Params["ueGroupId"]
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleRemoveEeGroupSubscriptions(msg.ResponseChan, ueGroupId, subsId)
				case message.EventUpdateEeGroupSubscriptions:
					// TODO
					ueGroupId := msg.HTTPRequest.Params["ueGroupId"]
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleUpdateEeGroupSubscriptions(msg.ResponseChan, ueGroupId, subsId, msg.HTTPRequest.Body.(models.EeSubscription))
				case message.EventCreateEeGroupSubscriptions:
					// TODO
					ueGroupId := msg.HTTPRequest.Params["ueGroupId"]
					producer.HandleCreateEeGroupSubscriptions(msg.ResponseChan, ueGroupId, msg.HTTPRequest.Body.(models.EeSubscription))
				case message.EventQueryEeGroupSubscriptions:
					// TODO
					ueGroupId := msg.HTTPRequest.Params["ueGroupId"]
					producer.HandleQueryEeGroupSubscriptions(msg.ResponseChan, ueGroupId)
				case message.EventRemoveeeSubscriptions:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleRemoveeeSubscriptions(msg.ResponseChan, ueId, subsId)
				case message.EventUpdateEesubscriptions:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleUpdateEesubscriptions(msg.ResponseChan, ueId, subsId, msg.HTTPRequest.Body.(models.EeSubscription))
				case message.EventCreateEeSubscriptions:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleCreateEeSubscriptions(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.EeSubscription))
				case message.EventQueryeesubscriptions:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQueryeesubscriptions(msg.ResponseChan, ueId)
				case message.EventPatchOperSpecData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandlePatchOperSpecData(msg.ResponseChan, ueId, msg.HTTPRequest.Body.([]models.PatchItem))
				case message.EventQueryOperSpecData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQueryOperSpecData(msg.ResponseChan, ueId)
				case message.EventGetppData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleGetppData(msg.ResponseChan, ueId)
				case message.EventCreateSessionManagementData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					pduSessionId, _ := strconv.ParseInt(msg.HTTPRequest.Params["pduSessionId"], 10, 64)
					producer.HandleCreateSessionManagementData(msg.ResponseChan, ueId, int32(pduSessionId), msg.HTTPRequest.Body.(models.PduSessionManagementData))
				case message.EventDeleteSessionManagementData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					pduSessionId, _ := strconv.ParseInt(msg.HTTPRequest.Params["pduSessionId"], 10, 64)
					producer.HandleDeleteSessionManagementData(msg.ResponseChan, ueId, int32(pduSessionId))
				case message.EventQuerySessionManagementData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					pduSessionId, _ := strconv.ParseInt(msg.HTTPRequest.Params["pduSessionId"], 10, 64)
					producer.HandleQuerySessionManagementData(msg.ResponseChan, ueId, int32(pduSessionId))
				case message.EventQueryProvisionedData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					servingPlmnId := msg.HTTPRequest.Params["servingPlmnId"]
					producer.HandleQueryProvisionedData(msg.ResponseChan, ueId, servingPlmnId)
				case message.EventModifyPpData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleModifyPpData(msg.ResponseChan, ueId, msg.HTTPRequest.Body.([]models.PatchItem))
				case message.EventGetAmfSubscriptionInfo:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleGetAmfSubscriptionInfo(msg.ResponseChan, ueId, subsId)
				case message.EventGetIdentityData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleGetIdentityData(msg.ResponseChan, ueId)
				case message.EventGetOdbData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleGetOdbData(msg.ResponseChan, ueId)
				case message.EventGetSharedData:
					// TODO
					var sharedDataIds []string
					if len(msg.HTTPRequest.Query["shared-data-ids"]) != 0 {
						sharedDataIds = msg.HTTPRequest.Query["shared-data-ids"]
						if strings.Contains(sharedDataIds[0], ",") {
							sharedDataIds = strings.Split(sharedDataIds[0], ",")
						}
					}
					producer.HandleGetSharedData(msg.ResponseChan, sharedDataIds)
				case message.EventRemovesdmSubscriptions:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleRemovesdmSubscriptions(msg.ResponseChan, ueId, subsId)
				case message.EventUpdatesdmsubscriptions:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleUpdatesdmsubscriptions(msg.ResponseChan, ueId, subsId, msg.HTTPRequest.Body.(models.SdmSubscription))
				case message.EventCreateSdmSubscriptions:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleCreateSdmSubscriptions(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.SdmSubscription))
				case message.EventQuerysdmsubscriptions:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQuerysdmsubscriptions(msg.ResponseChan, ueId)
				case message.EventQuerySmData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					servingPlmnId := msg.HTTPRequest.Params["servingPlmnId"]

					singleNssai := models.Snssai{}
					singleNssaiQuery := msg.HTTPRequest.Query.Get("single-nssai")
					err := json.Unmarshal([]byte(singleNssaiQuery), &singleNssai)
					if err != nil {
						HandlerLog.Warnln(err)
					}

					dnn := msg.HTTPRequest.Query.Get("dnn")

					producer.HandleQuerySmData(msg.ResponseChan, ueId, servingPlmnId, singleNssai, dnn)
				case message.EventCreateSmfContextNon3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					pduSessionId, _ := strconv.ParseInt(msg.HTTPRequest.Params["pduSessionId"], 10, 64)
					producer.HandleCreateSmfContextNon3gpp(msg.ResponseChan, ueId, int32(pduSessionId), msg.HTTPRequest.Body.(models.SmfRegistration))
				case message.EventDeleteSmfContext:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					pduSessionId := msg.HTTPRequest.Params["pduSessionId"]
					producer.HandleDeleteSmfContext(msg.ResponseChan, ueId, pduSessionId)
				case message.EventQuerySmfRegistration:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					pduSessionId := msg.HTTPRequest.Params["pduSessionId"]
					producer.HandleQuerySmfRegistration(msg.ResponseChan, ueId, pduSessionId)
				case message.EventQuerySmfRegList:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQuerySmfRegList(msg.ResponseChan, ueId)
				case message.EventQuerySmfSelectData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					servingPlmnId := msg.HTTPRequest.Params["servingPlmnId"]
					producer.HandleQuerySmfSelectData(msg.ResponseChan, ueId, servingPlmnId)
				case message.EventCreateSmsfContext3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleCreateSmsfContext3gpp(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.SmsfRegistration))
				case message.EventDeleteSmsfContext3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleDeleteSmsfContext3gpp(msg.ResponseChan, ueId)
				case message.EventQuerySmsfContext3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQuerySmsfContext3gpp(msg.ResponseChan, ueId)
				case message.EventCreateSmsfContextNon3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleCreateSmsfContextNon3gpp(msg.ResponseChan, ueId, msg.HTTPRequest.Body.(models.SmsfRegistration))
				case message.EventDeleteSmsfContextNon3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleDeleteSmsfContextNon3gpp(msg.ResponseChan, ueId)
				case message.EventQuerySmsfContextNon3gpp:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					producer.HandleQuerySmsfContextNon3gpp(msg.ResponseChan, ueId)
				case message.EventQuerySmsMngData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					servingPlmnId := msg.HTTPRequest.Params["servingPlmnId"]
					producer.HandleQuerySmsMngData(msg.ResponseChan, ueId, servingPlmnId)
				case message.EventQuerySmsData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					servingPlmnId := msg.HTTPRequest.Params["servingPlmnId"]
					producer.HandleQuerySmsData(msg.ResponseChan, ueId, servingPlmnId)
				case message.EventPostSubscriptionDataSubscriptions:
					// TODO
					producer.HandlePostSubscriptionDataSubscriptions(msg.ResponseChan, msg.HTTPRequest.Body.(models.SubscriptionDataSubscriptions))
				case message.EventRemovesubscriptionDataSubscriptions:
					// TODO
					subsId := msg.HTTPRequest.Params["subsId"]
					producer.HandleRemovesubscriptionDataSubscriptions(msg.ResponseChan, subsId)
				case message.EventQueryTraceData:
					// TODO
					ueId := msg.HTTPRequest.Params["ueId"]
					servingPlmnId := msg.HTTPRequest.Params["servingPlmnId"]
					producer.HandleQueryTraceData(msg.ResponseChan, ueId, servingPlmnId)
				default:
					HandlerLog.Warnf("Event[%d] has not been implemented", msg.Event)
				}
			} else {
				HandlerLog.Errorln("Channel closed!")
			}

		case <-time.After(time.Second * 1):

		}
	}
}
