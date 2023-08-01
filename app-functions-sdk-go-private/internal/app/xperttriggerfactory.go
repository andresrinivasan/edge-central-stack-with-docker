// Copyright (C) 2021-2023 IOTech Ltd

package app

import (
	"errors"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/mqtt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
)

func (svc *Service) setupTriggers() (map[string]interfaces.Trigger, error) {
	result := make(map[string]interfaces.Trigger)

	bnd := NewTriggerServiceBinding(svc)
	mp := NewTriggerMessageProcessor(bnd, svc.MetricsManager())

	triggerTypes := svc.config.Trigger.Types()

	if len(triggerTypes) == 0 {
		return nil, errors.New("no Trigger Type defined")
	}

	for _, triggerType := range triggerTypes {
		switch triggerType {
		case internal.TriggerTypeHTTP:
			svc.LoggingClient().Info("HTTP trigger selected")
			result[triggerType] = http.NewTrigger(bnd, mp, svc.webserver)
		case internal.TriggerTypeMessageBus:
			svc.LoggingClient().Info("EdgeX MessageBus trigger selected")
			result[triggerType] = messagebus.NewTrigger(bnd, mp)
		case internal.TriggerTypeMQTT:
			svc.LoggingClient().Info("External MQTT trigger selected")
			result[triggerType] = mqtt.NewTrigger(bnd, mp)
		default:
			if factory, found := svc.customTriggerFactories[triggerType]; found {
				t, err := factory(svc)
				if err != nil {
					svc.LoggingClient().Errorf("failed to initialize custom trigger [%s]: %s", t, err.Error())
					return nil, err
				}
				result[triggerType] = t
			} else {
				return nil, fmt.Errorf("custom trigger [%s] is not registered", triggerType)
			}
		}
	}

	return result, nil
}
