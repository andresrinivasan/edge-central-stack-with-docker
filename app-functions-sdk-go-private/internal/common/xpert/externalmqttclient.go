// Copyright (C) 2021 IOTech Ltd

package xpert

import (
	"time"

	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
)

type SharedMQTTClient struct {
	client               pahoMQTT.Client
	secretsLastRetrieved time.Time
	clientOptions        *pahoMQTT.ClientOptions
}

func (c *SharedMQTTClient) Get() (pahoMQTT.Client, time.Time) {
	return c.client, c.secretsLastRetrieved
}

func (c *SharedMQTTClient) Set(client pahoMQTT.Client, secretsLastRetrieved time.Time) {
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(0)
	}
	c.client = client
	c.secretsLastRetrieved = secretsLastRetrieved
}

func (c *SharedMQTTClient) GetClientOptions() *pahoMQTT.ClientOptions {
	return c.clientOptions
}

func (c *SharedMQTTClient) SetClientOptions(opts *pahoMQTT.ClientOptions) {
	c.clientOptions = opts
}
