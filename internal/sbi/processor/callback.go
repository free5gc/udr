package processor

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/free5gc/openapi/Nudr_DataRepository"
	"github.com/free5gc/openapi/models"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/internal/util"
)

var CurrentResourceUri string

func PreHandleOnDataChangeNotify(ueId string, resourceId string, patchItems []models.PatchItem,
	origValue interface{}, newValue interface{},
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
	policyDataChangeNotification := models.PolicyDataChangeNotification{}

	if ueId != "" {
		policyDataChangeNotification.UeId = ueId
	}

	switch v := value.(type) {
	case models.AmPolicyData:
		policyDataChangeNotification.AmPolicyData = &v
	case models.UePolicySet:
		policyDataChangeNotification.UePolicySet = &v
	case models.SmPolicyData:
		policyDataChangeNotification.SmPolicyData = &v
	case models.UsageMonData:
		policyDataChangeNotification.UsageMonId = dataId
		policyDataChangeNotification.UsageMonData = &v
	case models.SponsorConnectivityData:
		policyDataChangeNotification.SponsorId = dataId
		policyDataChangeNotification.SponsorConnectivityData = &v
	case models.BdtData:
		policyDataChangeNotification.BdtRefId = dataId
		policyDataChangeNotification.BdtData = &v
	default:
		return
	}

	go SendPolicyDataChangeNotification(policyDataChangeNotification)
}

func PreHandleInfluenceDataUpdateNotification(influenceId string, original, modified *models.TrafficInfluData) {
	resUri := fmt.Sprintf("%s/application-data/influenceData/%s",
		udr_context.GetSelf().GetIPv4GroupUri(udr_context.NUDR_DR), influenceId)

	go SendInfluenceDataUpdateNotification(resUri, original, modified)
}

func SendOnDataChangeNotify(ueId string, notifyItems []models.NotifyItem) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.HttpLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	udrSelf := udr_context.GetSelf()
	configuration := Nudr_DataRepository.NewConfiguration()
	client := Nudr_DataRepository.NewAPIClient(configuration)

	for _, subscriptionDataSubscription := range udrSelf.SubscriptionDataSubscriptions {
		if ueId == subscriptionDataSubscription.UeId {
			onDataChangeNotifyUrl := subscriptionDataSubscription.CallbackReference

			dataChangeNotify := models.DataChangeNotify{}
			dataChangeNotify.UeId = ueId
			dataChangeNotify.OriginalCallbackReference = []string{subscriptionDataSubscription.OriginalCallbackReference}
			dataChangeNotify.NotifyItems = notifyItems
			httpResponse, err := client.DataChangeNotifyCallbackDocumentApi.OnDataChangeNotify(
				context.TODO(), onDataChangeNotifyUrl, dataChangeNotify)
			if err != nil {
				logger.HttpLog.Errorln(err.Error())
			} else if httpResponse == nil {
				logger.HttpLog.Errorln("Empty HTTP response")
			}

			defer func() {
				if httpResponse.Body != nil {
					if err := httpResponse.Body.Close(); err != nil {
						logger.HttpLog.Errorln("Failed to close response body:", err)
					}
				}
			}()
		}
	}
}

func SendPolicyDataChangeNotification(policyDataChangeNotification models.PolicyDataChangeNotification) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.HttpLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	udrSelf := udr_context.GetSelf()

	for _, policyDataSubscription := range udrSelf.PolicyDataSubscriptions {
		policyDataChangeNotificationUrl := policyDataSubscription.NotificationUri

		configuration := Nudr_DataRepository.NewConfiguration()
		client := Nudr_DataRepository.NewAPIClient(configuration)
		httpResponse, err := client.PolicyDataChangeNotifyCallbackDocumentApi.PolicyDataChangeNotify(
			context.TODO(), policyDataChangeNotificationUrl, policyDataChangeNotification)
		if err != nil {
			logger.HttpLog.Errorln(err.Error())
		} else if httpResponse == nil {
			logger.HttpLog.Errorln("Empty HTTP response")
		}

		defer func() {
			if httpResponse.Body != nil {
				if err := httpResponse.Body.Close(); err != nil {
					logger.HttpLog.Errorln("Failed to close response body:", err)
				}
			}
		}()
	}
}

func SendInfluenceDataUpdateNotification(resUri string, original, modified *models.TrafficInfluData) {
	udrSelf := udr_context.GetSelf()

	configuration := Nudr_DataRepository.NewConfiguration()
	client := Nudr_DataRepository.NewAPIClient(configuration)

	var trafficInfluDataNotif models.TrafficInfluDataNotif
	trafficInfluDataNotif.ResUri = resUri
	udrSelf.InfluenceDataSubscriptions.Range(func(key, value interface{}) bool {
		influenceDataSubscription, ok := value.(*models.TrafficInfluSub)
		if !ok {
			logger.HttpLog.Errorf("Failed to load influenceData subscription ID [%+v]", key)
			return true
		}
		influenceDataChangeNotificationUrl := influenceDataSubscription.NotificationUri

		// Check if the modified data is subscribed
		// If positive, send notification about the update
		if checkInfluenceDataSubscription(modified, influenceDataSubscription) {
			logger.HttpLog.Tracef("Send notification about update of influence data")
			trafficInfluDataNotif.TrafficInfluData = modified
			httpResponse, err := client.InfluenceDataUpdateNotifyCallbackDocumentApi.InfluenceDataChangeNotify(context.TODO(),
				influenceDataChangeNotificationUrl, []models.TrafficInfluDataNotif{trafficInfluDataNotif})
			if err != nil {
				logger.HttpLog.Errorln(err.Error())
			} else if httpResponse == nil {
				logger.HttpLog.Errorln("Empty HTTP response")
			} else {
				defer func() {
					if httpResponse.Body != nil {
						if err := httpResponse.Body.Close(); err != nil {
							logger.HttpLog.Errorln("Failed to close response body:", err)
						}
					}
				}()
			}
		} else if checkInfluenceDataSubscription(original, influenceDataSubscription) {
			// If the modified data is not subscribed or nil, check if the original data is subscribed
			// If positive, send notification about the removal
			logger.HttpLog.Tracef("Send notification about removal of influence data")
			trafficInfluDataNotif.TrafficInfluData = nil
			httpResponse, err := client.InfluenceDataUpdateNotifyCallbackDocumentApi.InfluenceDataChangeNotify(context.TODO(),
				influenceDataChangeNotificationUrl, []models.TrafficInfluDataNotif{trafficInfluDataNotif})
			if err != nil {
				logger.HttpLog.Errorln(err.Error())
			} else if httpResponse == nil {
				logger.HttpLog.Errorln("Empty HTTP response")
			} else {
				defer func() {
					if httpResponse.Body != nil {
						if err := httpResponse.Body.Close(); err != nil {
							logger.HttpLog.Errorln("Failed to close response body:", err)
						}
					}
				}()
			}
		}
		return true
	})
}

func checkInfluenceDataSubscription(data *models.TrafficInfluData, sub *models.TrafficInfluSub) bool {
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
