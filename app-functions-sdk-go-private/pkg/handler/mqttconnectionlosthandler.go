package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	PostDisconnectionAlert = "PostDisconnectionAlert"
	ServiceKey             = "ServiceKey"
)

func MqttConnectionLostHandler(lc logger.LoggingClient, nc interfaces.NotificationClient,
	serviceKey string, postDisconnectionAlert string) func(mqttClient MQTT.Client, err error) {
	return func(mqttClient MQTT.Client, err error) {
		mqttOptions := mqttClient.OptionsReader()
		var servers []string
		for _, url := range mqttOptions.Servers() {
			servers = append(servers, url.String())
		}
		notificationContent := dtos.NewDisconnectionNotificationContent(
			servers,
			mqttOptions.ClientID(),
			err.Error(),
			time.Now().Format(time.RFC1123), // DAY, DD MON YYYY hh:mm:ss TimeZone
		)

		msg, err := notificationContent.JsonString()
		if err != nil {
			lc.Errorf("failed to parse notification content, err: %v", err)
			return
		}
		lc.Errorf("MQTT connection lost: %s", msg)

		isContinue, err := strconv.ParseBool(postDisconnectionAlert)
		if err != nil {
			lc.Warnf("invalid value for application setting %s. Error: %v. Use default value: %s",
				PostDisconnectionAlert, err, common.ValueFalse)
		}
		if !isContinue {
			return
		}
		go sendNotification(lc, nc, serviceKey, msg)
	}
}

func sendNotification(lc logger.LoggingClient, nc interfaces.NotificationClient, serviceKey string, content string) {
	if nc == nil {
		lc.Error("support-notifications client is not ready. Please check the [Clients] section of the " +
			"service configuration.")
		return
	}

	dto := dtos.Notification{
		Content:     content,
		ContentType: common.ContentTypeJSON,
		Category:    common.DisconnectAlert,
		Labels:      []string{common.MQTT},
		Sender:      serviceKey,
		Severity:    models.Critical,
	}
	req := requests.NewAddNotificationRequest(dto)
	res, err := nc.SendNotification(context.Background(), []requests.AddNotificationRequest{req})
	if err != nil {
		lc.Errorf("fail to send a notification to support-notifications service, err: %v", err)
	}
	if len(res) > 0 && res[0].StatusCode > http.StatusMultiStatus {
		lc.Errorf("failed to create a notification on support-notifications service, err: %v", res[0].Message)
	}
}
