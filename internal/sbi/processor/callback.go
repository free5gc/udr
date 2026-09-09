package processor

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/free5gc/openapi/models"
	DataRepository "github.com/free5gc/openapi/udr/DR"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/internal/util"
)

var CurrentResourceUri string

func getCallbackTokenCtx(
	serviceName models.Nrf_NFMgmt_ServiceName,
	targetNF models.Nrf_NFMgmt_NFType,
	targetNFInstanceID string,
	operation string,
) (context.Context, bool) {
	ctx, pd, err := udr_context.GetSelf().GetTokenCtxForNFInstance(
		serviceName, targetNF, targetNFInstanceID)
	if err != nil {
		logger.SBILog.Errorf("%s get token failed: %v", operation, err)
		if pd != nil {
			logger.SBILog.Errorf("%s get token problem details: %+v", operation, pd)
		}
		return nil, false
	}
	if pd != nil {
		logger.SBILog.Errorf("%s get token problem details: %+v", operation, pd)
		return nil, false
	}
	return ctx, true
}

func PreHandleOnDataChangeNotify(ueId string, resourceId string, patchItems []models.PatchItem,
	origValue map[string]interface{}, newValue map[string]interface{},
) {
	notifyItems := []models.NotifyItem{}
	changes := []models.ChangeItem{}

	for _, patchItem := range patchItems {
		change := models.ChangeItem{
			Op:        models.ChangeType(patchItem.Op),
			Path:      patchItem.Path,
			From:      patchItem.From,
			OrigValue: origValue,
			NewValue:  newValue,
		}
		changes = append(changes, change)
	}

	notifyItem := models.NotifyItem{
		ResourceId: resourceId,
		Changes:    changes,
	}

	notifyItems = append(notifyItems, notifyItem)

	go SendOnDataChangeNotify(ueId, notifyItems)
}

func PreHandlePolicyDataChangeNotification(ueId string, dataId string, value interface{}) {
	policyDataChangeNotification := models.Udr_DR_PolicyDataChangeNotification{}

	if ueId != "" {
		policyDataChangeNotification.UeId = ueId
	}

	switch v := value.(type) {
	case models.Udr_DR_AmPolicyData:
		policyDataChangeNotification.AmPolicyData = &v
	case models.Udr_DR_UePolicySet:
		policyDataChangeNotification.UePolicySet = &v
	case models.Udr_DR_SmPolicyData:
		policyDataChangeNotification.SmPolicyData = &v
	case models.Udr_DR_UsageMonData:
		policyDataChangeNotification.UsageMonId = dataId
		policyDataChangeNotification.UsageMonData = &v
	case models.Udr_DR_SponsorConnectivityData:
		policyDataChangeNotification.SponsorId = dataId
		policyDataChangeNotification.SponsorConnectivityData = &v
	case models.Udr_DR_BdtData:
		policyDataChangeNotification.BdtRefId = dataId
		policyDataChangeNotification.BdtData = &v
	default:
		return
	}

	go SendPolicyDataChangeNotification(policyDataChangeNotification)
}

func PreHandleInfluenceDataUpdateNotification(influenceId string, original, modified *models.Udr_DR_TrafficInfluData) {
	resUri := fmt.Sprintf("%s/application-data/influenceData/%s",
		udr_context.GetSelf().GetIPv4GroupUri(udr_context.NUDR_DR), influenceId)

	go SendInfluenceDataUpdateNotification(resUri, original, modified)
}

func SendOnDataChangeNotify(ueId string, notifyItems []models.NotifyItem) {
	defer func() {
		if p := recover(); p != nil {
			logger.HttpLog.Errorf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	udrSelf := udr_context.GetSelf()
	configuration := DataRepository.NewConfiguration()
	client := DataRepository.NewAPIClient(configuration)

	for _, record := range udrSelf.SubscriptionDataSubscriptions {
		if record == nil || record.Subscription == nil {
			logger.SBILog.Errorln("Invalid subscription data subscription record")
			continue
		}
		subscriptionDataSubscription := record.Subscription
		if ueId == subscriptionDataSubscription.UeId {
			ctx, ok := getCallbackTokenCtx(
				record.CallbackTarget.ServiceName,
				record.CallbackTarget.NfType,
				record.CallbackTarget.NfInstanceID,
				"SendOnDataChangeNotify",
			)
			if !ok {
				continue
			}

			onDataChangeNotifyUrl := subscriptionDataSubscription.CallbackReference

			dataChangeReq := DataRepository.SubscriptionDataSubscriptionsOnDataChangeRequest{
				RequestBody: &models.Udr_DR_DataChangeNotify{
					UeId: ueId,
					OriginalCallbackReference: []string{
						subscriptionDataSubscription.OriginalCallbackReference,
					},
					NotifyItems: notifyItems,
				},
			}
			rsp, err := client.SubsToNotifyCollectionApi.SubscriptionDataSubscriptionsOnDataChange(
				ctx, onDataChangeNotifyUrl, &dataChangeReq)

			if err != nil {
				logger.SBILog.Errorln(err.Error())
			} else if rsp == nil {
				logger.SBILog.Errorln("Empty SubscriptionDataSubscriptionsOnDataChangePost response")
			}
		}
	}
}

func SendPolicyDataChangeNotification(policyDataChangeNotification models.Udr_DR_PolicyDataChangeNotification) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.HttpLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	udrSelf := udr_context.GetSelf()

	for _, record := range udrSelf.PolicyDataSubscriptions {
		if record == nil || record.Subscription == nil {
			logger.SBILog.Errorln("Invalid policy data subscription record")
			continue
		}
		ctx, ok := getCallbackTokenCtx(
			record.CallbackTarget.ServiceName,
			record.CallbackTarget.NfType,
			record.CallbackTarget.NfInstanceID,
			"SendPolicyDataChangeNotification",
		)
		if !ok {
			continue
		}

		policyDataSubscription := record.Subscription
		policyDataChangeNotificationUrl := policyDataSubscription.NotificationUri

		configuration := DataRepository.NewConfiguration()
		client := DataRepository.NewAPIClient(configuration)

		req := DataRepository.CreateIndividualPolicyDataSubscriptionPolicyDataChangeNotificationRequest{
			RequestBody: []models.Udr_DR_PolicyDataChangeNotification{
				policyDataChangeNotification,
			},
		}
		rsp, err := client.PolicyDataSubscriptionsCollectionApi.
			CreateIndividualPolicyDataSubscriptionPolicyDataChangeNotification(ctx,
				policyDataChangeNotificationUrl, &req)

		if err != nil {
			logger.SBILog.Errorln(err.Error())
		} else if rsp == nil {
			logger.SBILog.Errorln("Empty CreateIndividualPolicyDataSubscriptionPolicyDataChangeNotificationPost response")
		}
	}
}

func SendInfluenceDataUpdateNotification(resUri string, original, modified *models.Udr_DR_TrafficInfluData) {
	udrSelf := udr_context.GetSelf()

	configuration := DataRepository.NewConfiguration()
	client := DataRepository.NewAPIClient(configuration)

	var trafficInfluDataNotif models.Udr_DR_TrafficInfluDataNotif
	trafficInfluDataNotif.ResUri = resUri
	udrSelf.InfluenceDataSubscriptions.Range(func(key, value interface{}) bool {
		record, ok := value.(*udr_context.InfluenceDataSubscriptionRecord)
		if !ok || record.Subscription == nil {
			logger.HttpLog.Errorf("Failed to load influenceData subscription ID [%+v]", key)
			return true
		}
		influenceDataSubscription := record.Subscription
		ctx, ok := getCallbackTokenCtx(
			record.CallbackTarget.ServiceName,
			record.CallbackTarget.NfType,
			record.CallbackTarget.NfInstanceID,
			"SendInfluenceDataUpdateNotification",
		)
		if !ok {
			return true
		}
		influenceDataChangeNotificationUrl := influenceDataSubscription.NotificationUri

		// Check if the modified data is subscribed
		// If positive, send notification about the update
		if checkInfluenceDataSubscription(modified, influenceDataSubscription) {
			logger.HttpLog.Tracef("Send notification about update of influence data")
			trafficInfluDataNotif.TrafficInfluData = modified

			req := DataRepository.CreateIndividualInfluenceDataSubscriptionTrafficInfluenceDataChangeNotificationRequest{
				RequestBody: []interface{}{trafficInfluDataNotif},
			}
			rsp, err := client.InfluenceDataSubscriptionsCollectionApi.
				CreateIndividualInfluenceDataSubscriptionTrafficInfluenceDataChangeNotification(
					ctx, influenceDataChangeNotificationUrl, &req)

			if err != nil {
				logger.SBILog.Errorln(err.Error())
			} else if rsp == nil {
				logger.SBILog.Errorln(
					"Empty CreateIndividualInfluenceDataSubscriptionTrafficInfluenceDataChangeNotificationPost response")
			}
		} else if checkInfluenceDataSubscription(original, influenceDataSubscription) {
			// If the modified data is not subscribed or nil, check if the original data is subscribed
			// If positive, send notification about the removal
			logger.HttpLog.Tracef("Send notification about removal of influence data")
			trafficInfluDataNotif.TrafficInfluData = nil
			req := DataRepository.CreateIndividualInfluenceDataSubscriptionTrafficInfluenceDataChangeNotificationRequest{
				RequestBody: []interface{}{trafficInfluDataNotif},
			}
			rsp, err := client.InfluenceDataSubscriptionsCollectionApi.
				CreateIndividualInfluenceDataSubscriptionTrafficInfluenceDataChangeNotification(
					ctx, influenceDataChangeNotificationUrl, &req)

			if err != nil {
				logger.SBILog.Errorln(err.Error())
			} else if rsp == nil {
				logger.SBILog.Errorln(
					"Empty CreateIndividualInfluenceDataSubscriptionTrafficInfluenceDataChangeNotificationPost response")
			}
		}
		return true
	})
}

func checkInfluenceDataSubscription(data *models.Udr_DR_TrafficInfluData, sub *models.Udr_DR_TrafficInfluSub) bool {
	if data == nil || sub == nil {
		return false
	}
	if data.Dnn != "" && !util.Contain(data.Dnn, sub.Dnns) {
		return false
	} else if data.Snssai != nil && !util.Contain(*data.Snssai, sub.Snssais) {
		return false
	} else if data.InterGroupId != "AnyUE" {
		if data.InterGroupId != "" && !util.Contain(data.InterGroupId, sub.InternalGroupIds) {
			return false
		} else if data.Supi != "" && !util.Contain(data.Supi, sub.Supis) {
			return false
		}
	}
	return true
}
