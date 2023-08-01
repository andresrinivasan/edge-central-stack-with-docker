// Copyright (C) 2021 IOTech Ltd

package xpert

import (
	"errors"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

const (
	awsIoTThingCert = `-----BEGIN CERTIFICATE-----
MIIDWTCCAkGgAwIBAgIUNQ4fnr2HwjCCaOc9Fgv7EOxBIVcwDQYJKoZIhvcNAQEL
BQAwTTFLMEkGA1UECwxCQW1hem9uIFdlYiBTZXJ2aWNlcyBPPUFtYXpvbi5jb20g
SW5jLiBMPVNlYXR0bGUgU1Q9V2FzaGluZ3RvbiBDPVVTMB4XDTIwMDcwOTA0NDU1
N1oXDTQ5MTIzMTIzNTk1OVowHjEcMBoGA1UEAwwTQVdTIElvVCBDZXJ0aWZpY2F0
ZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANXZ2QyV5TrN1A6lXpME
whWosSkFO1OE1i/GGw+WxEinujNDteBIsf2ySkvk9hk00FLFiUjhiDRjamO2joDH
yrfE8GMxLuruxpSk9+czgw/ZC2eJ2FaJukED5pNxNovTEDRnoOdqx+jz7uYbnwPC
0DgBuGWEzWe9eAKAwjb2KBbMFPCo6diMYa0Xbh6oNrgs0x/nE80eOacF+DD2MRdn
MFgRo+pvDsNR5E+mma10VdvW7Q8SgqPTVGG2iv25CMKyu4uWUpUxlW0WhtxrTUbo
KLigD1Swkl9jze/Htv2HT+qCAmZgw8m/EuSJD6W1jJWcGIDklMDbx1fw0i8bujgp
f58CAwEAAaNgMF4wHwYDVR0jBBgwFoAUrImIwCDoAN2kokCeH4ehXhrXwDMwHQYD
VR0OBBYEFKxdXvgKH1J9LLkEsUyNg2RWP0PHMAwGA1UdEwEB/wQCMAAwDgYDVR0P
AQH/BAQDAgeAMA0GCSqGSIb3DQEBCwUAA4IBAQCSbPJreobWzipZbt7EzrT/1Cvt
/srQi+tOy/jxyYCIzqJAfL+ks/vrzPqUlJghb3dYKH+XnirCUFcRqbtDzA9AXV4F
UnO08a8JQZJ7uQ8+EUjg5Esks1EWy+D4OZv7FKCFoe1aDZeJklbAjdGNPXpgASjA
TWjew5Ya1hROF28D0G4ZEWmtiZ5qTuOxD7KGDjoSPW6SuGFojJKic9t8ucoaDIzm
7wU1iDhKlWZ6gXlcGy22DNyf1gs9o0XwijshTG/ApiAjf9ynpAikPuRIZCHOcYyo
YzS0Nabq/+jujMawzPB1tbqCvXG/gJVe829KLF/wMBoaSMH6PGs4QujTI72B
-----END CERTIFICATE-----`
	awsIoTThingPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA1dnZDJXlOs3UDqVekwTCFaixKQU7U4TWL8YbD5bESKe6M0O1
4Eix/bJKS+T2GTTQUsWJSOGINGNqY7aOgMfKt8TwYzEu6u7GlKT35zODD9kLZ4nY
Vom6QQPmk3E2i9MQNGeg52rH6PPu5hufA8LQOAG4ZYTNZ714AoDCNvYoFswU8Kjp
2IxhrRduHqg2uCzTH+cTzR45pwX4MPYxF2cwWBGj6m8Ow1HkT6aZrXRV29btDxKC
o9NUYbaK/bkIwrK7i5ZSlTGVbRaG3GtNRugouKAPVLCSX2PN78e2/YdP6oICZmDD
yb8S5IkPpbWMlZwYgOSUwNvHV/DSLxu6OCl/nwIDAQABAoIBAAHFl03KZCbTk8E3
T4xKSs1hI26ik3ZCsH+e1ZSQfDnZ4eoQ0o30LD1c3863K+7TiDJXXqbS74z3kecX
mSCqBxz0rcU0oB3kMpMCtuccfyZ/dt+5vagh5gAp5xwroPcRnZ3BigzAzr04YC9Z
5TxjvdPo970nl4gXgsoKhRFtgrGFpXkJwWJo9At6QxNFuczYNHrqaVYsRDKbf9Ri
S6gTMS/5T7R9ZRXhqKw/8GTBRMaR267ZrwY7Nzf/r/b/wrZwKw09D0eg6IGBgbTf
BO/yPv2WVpDjDPEw8H8ldWFZu8DN2bybVZxGc//XOko/0NOqLQzZTI8Q7QNKCZbU
qOfNc+ECgYEA9YYLWNTDQIAD0ufgWvpl0le8n1NoyGLsnPDiOTxhRT5D7Jz+MGhg
Bz95JI7xYXF0mPaL9rWv0BYYqyBW2QZj06kcCypPy6r5k2idN23cOoDZBnbfldCk
UzaXLasv4uRK76XeMbqpyNuK7Uznh7XT8aFOed//cUnw8koxGOxS3A8CgYEA3vnU
hwlIrfAdPwdB7lSj9lpeMjI8GpPq8qRwWq+4ed+UWZOj/GXopQDNcuf7ZaQTAUHX
OXc4DHHZS4EbmqmBl4utbFaFUdBANb/WqW3wna18ePd0qXez1xaPDPKVIzlRsPH6
sGqD857O7VGu3f0CvyRBDQi0slvEDqFhj6Jq03ECgYAFUOiv/LNZkywCBglCjwdj
XYj0/i5XoGS1JTYQvTDx+d4oomGSlL/3iDVMSFgLnxRCN5xiNB7hZ4kTM3kN6+h/
bbrwtvLRWxtaSLqWt6c8EQwh6rL+oGzebGErmPhJdl31AGdmNj903OQOLUsaEiLL
qY10cBgs0MgJxvd3La7BmwKBgDF/G+Jt+ShDaPqYzdXuDAefv9E8vYLY2wrJ3fcD
ktva+b94uqpIpQAb0X90Z6YEagOZbgFfqZ15mFbebhZDEnVlmDW4bxfeNqK31xr9
QLB/1mWz6L3FyLIyW8bwApMzIiM5VADdZDUsR5r+yuaUR4vOrHIMQLBnFnp48INF
9pjBAoGBAMcgXES4nf/1hh3r/P7pE3nkz+onJKSdb1jh3Ll10Jc9kGb+c8gtbR1Q
7xLagYHw05I0EXb5C6ArY2mReS+hnDenwC2EGC3AgDK25Ys1c6Pj54UDIobEwujU
HZXFBUZZYn65dnlzpCadBZPSp2q4SHK5u+VOcOT3gU9lVak1rbEK
-----END RSA PRIVATE KEY-----`
	awsIoTThingPrivateKeyMismatched = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA1dnZDJXlOs3UDqVekwTCFaixKQU7U4TWL8YbD5bESKe6M0O1
4Eix/bJKS+T2GTTQUsWJSOGINGNqY7aOgMfKt8TwYzEu6u7GlKT35zODD9kLZ4nY
Vom6QQPmk3E2i9MQNGeg52rH6PPu5hufA8LQOAG4ZYTNZ714AoDCNvYoFswU8Kjp
2IxhrRduHqg2uCzTH+cTzR45pwX4MPYxF2cwWBGj6m8Ow1HkT6aZrXRV29btDxKC
o9NUYbaK/bkIwrK7i5ZSlTGVbRaG3GtNRugouKAPVLCSX2PN78e2/YdP6oICZmDD
yb8S5IkPpbWMlZwYgOSUwNvHV/DSLxu6OCl/nwIDAQABAoIBAAHFl03KZCbTk8E3
T4xKSs1hI26ik3ZCsH+e1ZSQfDnZ4eoQ0o30LD1c3863K+7TiDJXXqbS74z3kecX
mSCqBxz0rcU0oB3kMpMCtuccfyZ/dt+5vagh5gAp5xwroPcRnZ3BigzAzr04YC9Z
5TxjvdPo970nl4gXgsoKhRFtgrGFpXkJwWJo9At6QxNFuczYNHrqaVYsRDKbf9Ri
S6gTMS/5T7R9ZRXhqKw/8GTBRMaR267ZrwY7Nzf/r/b/wrZwKw09D0eg6IGBgbTf
BO/yPv2WVpDjDPEw8H8ldWFZu8DN2bybVZxGc//XOko/0NOqLQzZTI8Q7QNKCZbU
qOfNc+ECgYEA9YYLWNTDQIAD0ufgWvpl0le8n1NoyGLsnPDiOTxhRT5D7Jz+MGhg
Bz95JI7xYXF0mPaL9rWv0BYYqyBW2QZj06kcCypPy6r5k2idN23cOoDZBnbfldCk
UzaXLasv4uRK76XeMbqpyNuK7Uznh7XT8aFOed//cUnw8koxGOxS3A8CgYEA3vnU
hwlIrfAdPwdB7lSj9lpeMjI8GpPq8qRwWq+4ed+UWZOj/GXopQDNcuf7ZaQTAUHX
OXc4DHHZS4EbmqmBl4utbFaFUdBANb/WqW3wna18ePd0qXez1xaPDPKVIzlRsPH6
sGqD857O7VGu3f0CvyRBDQi0slvEDqFhj6JqaaECgYAFUOiv/LNZkywCBglCjwdj
XYj0/i5XoGS1JTYQvTDx+d4oomGSlL/3iDVMSFgLnxRCN5xiNB7hZ4kTM3kN6+h/
bbrwtvLRWxtaSLqWt6c8EQwh6rL+oGzebGErmPhJdl31AGdmNj903OQOLUsaEiLL
qY10cBgs0MgJxvd3La7BmwKBgDF/G+Jt+ShDaPqYzdXuDAefv9E8vYLY2wrJ3fcD
ktva+b94uqpIpQAb0X90Z6YEagOZbgFfqZ15mFbebhZDEnVlmDW4bxfeNqK31xr9
QLB/1mWz6L3FyLIyW8bwApMzIiM5VADdZDUsR5r+yuaUR4vOrHIMQLBnFnp48INF
9pjBAoGBAMcgXES4nf/1hh3r/P7pE3nkz+onJKSdb1jh3Ll10Jc9kGb+c8gtbR1Q
7xLagYHw05I0EXb5C6ArY2mReS+hnDenwC2EGC3AgDK25Ys1c6Pj54UDIobEwujU
HZXFBUZZYn65dnlzpCadBZPSp2q4SHK5u+VOcOT3gU9lVak1rbEK
-----END RSA PRIVATE KEY-----`
)

func TestAWSIoTCoreMQTTSendNoParams(t *testing.T) {
	sender := NewAWSIoTCoreMQTTSender(AWSIoTCoreMQTTConfig{}, true, false)
	continuePipeline, result := sender.Send(ctx, []byte("data"))
	assert.False(t, continuePipeline)
	assert.Error(t, result.(error))
}

func TestAWSIoTCoreMQTTSendSetRetryDataPersistFalse(t *testing.T) {
	sender := NewAWSIoTCoreMQTTSender(AWSIoTCoreMQTTConfig{}, false, false)
	ctx.SetRetryData(nil)
	sender.setRetryData(ctx, []byte("data"))
	assert.Nil(t, ctx.RetryData())
}

func TestAWSIoTCoreMQTTSendSetRetryDataPersistTrue(t *testing.T) {
	sender := NewAWSIoTCoreMQTTSender(AWSIoTCoreMQTTConfig{}, true, false)
	ctx.SetRetryData(nil)
	sender.setRetryData(ctx, []byte("data"))
	assert.Equal(t, []byte("data"), ctx.RetryData())
}

func TestAWSIoTCoreMQTTClientFactoryValidateSecrets(t *testing.T) {
	factory := AWSIoTCoreMqttClientFactory{opts: MQTT.NewClientOptions()}
	tests := []struct {
		Name             string
		secrets          iotCoreSecrets
		ErrorExpectation bool
	}{
		{"No certificate", iotCoreSecrets{}, true},
		{"Mismatched key/cert", iotCoreSecrets{certPEMBlock: []byte(awsIoTThingCert), keyPEMBlock: []byte(awsIoTThingPrivateKeyMismatched)}, true},
		{"All satisfied", iotCoreSecrets{certPEMBlock: []byte(awsIoTThingCert), keyPEMBlock: []byte(awsIoTThingPrivateKey)}, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := factory.configureMQTTClientForAuth(test.secrets)
			if test.ErrorExpectation {
				assert.Error(t, result, "Result should be an error")
			} else {
				assert.Nil(t, result, "Should be nil")
				assert.NotNil(t, factory.opts.TLSConfig.Certificates)
			}
		})
	}
}

func TestAWSIoTCoreMQTTClientFactoryGetSecrets(t *testing.T) {
	factory := AWSIoTCoreMqttClientFactory{opts: MQTT.NewClientOptions()}
	mockSP := &mocks.SecretProvider{}
	mockSP.On("GetSecret", "/notfound").Return(nil, errors.New(""))
	mockSP.On("GetSecret", "/aws").Return(
		map[string]string{
			AWSIoTMQTTSecretClientCert: awsIoTThingCert,
			AWSIoTMQTTSecretClientKey:  awsIoTThingPrivateKey,
		}, nil)
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})

	tests := []struct {
		Name            string
		SecretPath      string
		ExpectedSecrets *iotCoreSecrets
		ExpectingError  bool
	}{
		{"No Secrets found", "/notfound", nil, true},
		{"With Secrets", "/aws", &iotCoreSecrets{
			certPEMBlock: []byte(awsIoTThingCert), keyPEMBlock: []byte(awsIoTThingPrivateKey)}, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			factory.secretPath = test.SecretPath
			secrets, err := factory.getSecrets(ctx)
			if test.ExpectingError {
				assert.Error(t, err, "Expecting error")
				return
			}
			assert.Equal(t, test.ExpectedSecrets, secrets)
		})
	}
}
