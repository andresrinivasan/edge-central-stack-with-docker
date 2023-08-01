// Copyright (C) 2021-2023 IOTech Ltd

package xpert

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common/xpert"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testIoTHubCert = `-----BEGIN CERTIFICATE-----
MIIDgzCCAmugAwIBAgIUXvDwsS+pW/W+hMRp675EcXsYwIgwDQYJKoZIhvcNAQEF
BQAwUTELMAkGA1UEBhMCVFcxEzARBgNVBAgMClNvbWUtU3RhdGUxDzANBgNVBAcM
BlRhaXBlaTEPMA0GA1UECgwGSU9UZWNoMQswCQYDVQQLDAJSRDAeFw0yMDA2MDIw
MzA3MjlaFw0zMDA1MzEwMzA3MjlaMFExCzAJBgNVBAYTAlRXMRMwEQYDVQQIDApT
b21lLVN0YXRlMQ8wDQYDVQQHDAZUYWlwZWkxDzANBgNVBAoMBklPVGVjaDELMAkG
A1UECwwCUkQwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDBW5VXLmNu
gcKAQ36CV/qGNSX2rxd4AubyJDp8e40zxMGmFPjNGpFgxmdk3iEEiQu8tLIqMWJZ
L6kKfAhc+gv2Wc/8PW6RP2Piuxi0e1sm/B/6Rnvnt6K7qqExsSGSbH6HHNoA2w9T
JvxA1daFY7AbzSms2MjGUFQhWbRMmes5/HuSWS9i7EBuv1oK2YmVy4CQSZLo56SA
9nTI++8fAdSDNtmSjsLRxOmUEjTh6Jxpz+eJHEz+a3a7bRB6+qsFcfew1m6raAMw
DH9Q5OhLkNdyN2/pIxQELMSDK//vVCfE5Y78sd2aY3wASUr6p/fsvmC31GI70LdG
1L7le8oGtSHZAgMBAAGjUzBRMB0GA1UdDgQWBBQxIH5o1gxpfmR5Njf4AAboLsJE
FzAfBgNVHSMEGDAWgBQxIH5o1gxpfmR5Njf4AAboLsJEFzAPBgNVHRMBAf8EBTAD
AQH/MA0GCSqGSIb3DQEBBQUAA4IBAQCcwd3XqTEObS3VYbSbVz71Rcoui/5Xc96J
9VyLaBJp+RrX0fNST5FYlyQAVrSKYtPsJH2/ffg2OvN+r4TZqfyNTx/rOoP+0VDl
l0ANV8XdXs9/Qbzgu+3XsOY3a6NCSTtV+PPgn8XOmO+sxmzafar2EHcacB4HcMlu
93adOxDQ5+42K35LcFgl2p1h3vOqWAyyPR8TFTgUfLxdl5KMczhGQgjRz2sLx9/3
rimceIOAN8dqF7rIeTZOAH0+TPx4Db79GuqVAQkWwYhnxyUQ1seFovd/fRpVMVHE
j7SAOX5YSoQ89JYP3hu4UgR/tOCUyGjpK2uiTREaDfjhQPuxQgz6
-----END CERTIFICATE-----
`
const testIoTHubKey = `-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDBW5VXLmNugcKA
Q36CV/qGNSX2rxd4AubyJDp8e40zxMGmFPjNGpFgxmdk3iEEiQu8tLIqMWJZL6kK
fAhc+gv2Wc/8PW6RP2Piuxi0e1sm/B/6Rnvnt6K7qqExsSGSbH6HHNoA2w9TJvxA
1daFY7AbzSms2MjGUFQhWbRMmes5/HuSWS9i7EBuv1oK2YmVy4CQSZLo56SA9nTI
++8fAdSDNtmSjsLRxOmUEjTh6Jxpz+eJHEz+a3a7bRB6+qsFcfew1m6raAMwDH9Q
5OhLkNdyN2/pIxQELMSDK//vVCfE5Y78sd2aY3wASUr6p/fsvmC31GI70LdG1L7l
e8oGtSHZAgMBAAECggEAPAZrWRIsiey8bccGKf6X5jvvmiIG3hnAiiEXCvfsAMsu
9pkCF+IMiYduJ5ERnO/SdfV+bbzA9EDocDnK+tohBowhFgQYI/0SzBsa80RsKrEQ
WEizRF7spOe2fM+pRWRq7xTU1Hksy7qJirMXknc8/5R0QJVX1sCvLV9EqpcJwAkH
t1L2laHXb4PelcZcj/I3EyNByeBlz00bEgrUl4+/28IwltFQ6TQKBCpWwiI1tq4l
rpMpIehec9ODHVgUAN2uOIh576HDVb4My130LMElX0fwgeQf3JXQ735e9XaQ3+j3
RMwL0PcB33GwZ4HSr5JF+K7R5qUuli7p5sqJjP5fFQKBgQDz4zGUdLhL6EEM69mt
E8ST3KLQuYOYhunBgjBO+3DR5heNkYUU3SI41kgmlTsmADNYP+5zMyymwWqZ9wnD
hHbvNdmGtPLch95oteRmjdZeNJq90dhdBzchYxgkU7KtMsKoJF6aKbhDtdIA0Gs5
cjUHq/5+d7IA+IyFS+wqLVA6hwKBgQDK9fNMjnD+gwEJaq/BDE3HZscsUhEquH5Y
QyKQ0g/6uv5ZsbdT2u3e59ZGXPkQE6v7WgeW9w2IdxwVI0bEFWNIH0zAZozeyhYS
wy63llPB5tyBphZ2ADj5faC+Tely8IkcWcZNNE+Hrv8jF6yxLGwPfKTsEKLEaBvO
Xyh3EuL4nwKBgFTY5ZbQRI2j732fT8t25RzL1Zjn8XBO/2Pi9wuDTmy3r9oAllv7
0rwTUGab5EgEKdi55SsO0qnxADUwTKVIoFf4VAUZTqSKYEXtgdhr3/hGNM91AeDb
ccKbxvpcY/z9e9sjTAY2HXTw/G5sE+GYafqRS6iT28marshw8Wh6+z5hAoGAPi0c
rM5SRVYCwkzBrOVFCpos2CIICktcwVNHyo/fv1L7yqSL4g+GoavqU8H1tvwfyq+o
9ZGXvr+mhb851aYrtROJosOH0lSccIEE1c8it5su4DTuWpX03lGjJcmeg8y2ZE4I
Vux4lLuCg9Cj4d8W96OarormIj82jYFPVzMc/0cCgYAn+hhEgBYlfpcEHE0CWXl2
BQoKn3pB+f0ImK8czQ2z4WcBt1s9+3aBbWELTjnxRgdKhp8S04a2TRv8n5LY8ayd
nV0wGkML7XQB8DCoRh9yWrzltV7IJs9NpqjKh8f6WfxzGiMfXXjARQR/XBaGBLVq
E0rqYET+TPPzUZSNSnphRw==
-----END PRIVATE KEY-----
`

func TestIoTHubMQTTSendSetRetryDataPersistFalse(t *testing.T) {
	sender := NewAzureIoTHubMQTTSender(AzureIoTHubMQTTConfig{}, false, false)
	ctx.SetRetryData(nil)
	sender.setRetryData(ctx, []byte("data"))
	assert.Nil(t, ctx.RetryData())
}

func TestIoTHubMQTTSendSetRetryDataPersistTrue(t *testing.T) {
	sender := NewAzureIoTHubMQTTSender(AzureIoTHubMQTTConfig{}, true, false)
	ctx.SetRetryData(nil)
	sender.setRetryData(ctx, []byte("data"))
	assert.Equal(t, []byte("data"), ctx.RetryData())
}

func TestIoTHubMQTTSendValidateSecrets(t *testing.T) {
	factory := IoTHubMqttClientFactory{opts: MQTT.NewClientOptions()}
	tests := []struct {
		Name             string
		secrets          iotHubSecrets
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{"No cert", iotHubSecrets{keyPEMBlock: []byte(testIoTHubKey)}, true, "not found"},
		{"No key", iotHubSecrets{certPEMBlock: []byte(testIoTHubCert)}, true, "not found"},
		{"All satisfied", iotHubSecrets{keyPEMBlock: []byte(testIoTHubKey),
			certPEMBlock: []byte(testIoTHubCert), username: "username"}, false, ""},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := factory.configureMQTTClientForAuth(test.secrets)
			if test.ErrorExpectation {
				assert.Error(t, result, "Result should be an error")
				assert.Contains(t, result.Error(), test.ErrorMessage)
			} else {
				assert.Nil(t, result, "Should be nil")
				assert.Len(t, factory.opts.TLSConfig.Certificates, 1)
			}
		})
	}
}

func TestIoTHubMQTTSendGetSecrets(t *testing.T) {
	factory := IoTHubMqttClientFactory{opts: MQTT.NewClientOptions()}
	mockSP := &mocks.SecretProvider{}
	mockSP.On("GetSecret", "notfound").Return(nil, errors.New(""))
	mockSP.On("GetSecret", "iothub").Return(
		map[string]string{
			IoTHubSecretClientKey:  testIoTHubKey,
			IoTHubSecretClientCert: testIoTHubCert,
		}, nil)
	mockSP.On("StoreSecrets").Return(nil)
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})

	tests := []struct {
		Name            string
		SecretPath      string
		ExpectedSecrets *iotHubSecrets
		ExpectingError  bool
	}{
		{"No Secrets found", "notfound", nil, true},
		{"With Secrets", "iothub", &iotHubSecrets{
			keyPEMBlock:  []byte(testIoTHubKey),
			certPEMBlock: []byte(testIoTHubCert),
		}, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			factory.secretPath = test.SecretPath
			iothubSecrets, err := factory.getSecrets(ctx)
			if test.ExpectingError {
				assert.Error(t, err, "Expecting error")
				return
			}
			require.Equal(t, test.ExpectedSecrets, iothubSecrets)
		})
	}
}

func TestIoTHubMQTTSendInvalidBrokerAddress(t *testing.T) {
	ctx.SetSharedMQTTClientMutex(&sync.Mutex{})
	ctx.SetMqttConnectionWaitingCounter(xpert.NewCounter(10, 0))

	mockSP := &mocks.SecretProvider{}
	mockSP.On("GetSecret", "iothub").Return(
		map[string]string{
			IoTHubSecretClientKey:  testIoTHubKey,
			IoTHubSecretClientCert: testIoTHubCert,
			IoTHubSecretUsername:   "username",
		}, nil)
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})

	iotHubConfig := AzureIoTHubMQTTConfig{
		SecretPath: "iothub",
		BaseMqttConfig: BaseMqttConfig{
			BrokerAddress: "aaa://unreachable:1111",
			Topic:         "devices/TestDevice01/messages/events/",
			ClientId:      "TestDevice01",
		},
	}

	sender := NewAzureIoTHubMQTTSender(iotHubConfig, false, false)
	assert.NoError(t, sender.initializeMQTTClient(ctx))
	result := sender.connect(ctx)
	require.Error(t, result, "Result should be an error")
	assert.True(t, strings.Contains(result.Error(), "not connect"), "Shall fail to access to AzureIOTHub.")
}

func TestIoTHubMQTTSendNoDataPassed(t *testing.T) {
	sender := NewAzureIoTHubMQTTSender(AzureIoTHubMQTTConfig{}, false, false)
	continuePipeline, result := sender.Send(ctx, nil)

	assert.False(t, continuePipeline, "Pipeline should stop")
	assert.Error(t, result.(error), "Result should be an error")
	assert.Equal(t, "no event received", result.(error).Error())
}
