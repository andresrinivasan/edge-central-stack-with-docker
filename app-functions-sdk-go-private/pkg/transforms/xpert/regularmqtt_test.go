// Copyright (C) 2021 IOTech Ltd

package xpert

import (
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	deviceCert = `-----BEGIN CERTIFICATE-----
MIIEDTCCAvWgAwIBAgIJAJPteGFrFeuoMA0GCSqGSIb3DQEBBQUAMIGcMQswCQYD
VQQGEwJUVzEVMBMGA1UECAwMVGFpd2FuIFIuTy5DMQ8wDQYDVQQHDAZUYWlwZWkx
DzANBgNVBAoMBklPVGVjaDEcMBoGA1UECwwTSW5mb3JtYXRpb24gU2VjdGlvbjES
MBAGA1UEAwwJRmVsaXhUaW5nMSIwIAYJKoZIhvcNAQkBFhNmZWxpeEBpb3RlY2hz
eXMuY29tMB4XDTE4MDkxODEwMDI1MVoXDTE5MDkxODEwMDI1MVowgZwxCzAJBgNV
BAYTAlRXMRUwEwYDVQQIDAxUYWl3YW4gUi5PLkMxDzANBgNVBAcMBlRhaXBlaTEP
MA0GA1UECgwGSU9UZWNoMRwwGgYDVQQLDBNJbmZvcm1hdGlvbiBTZWN0aW9uMRIw
EAYDVQQDDAlGZWxpeFRpbmcxIjAgBgkqhkiG9w0BCQEWE2ZlbGl4QGlvdGVjaHN5
cy5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC2Nl4umpeykVae
f1XlH70gdpsAXT47kyTPrLzcLFQviDE2CNtOEjRUlj3Yu2yXkk5YLJz6AxnUy/Yr
UtzuldfzHJNs+lVtJ/BN3M4IZqsgW8VEfVO6H0R4qKng0SN0/dZWLwjx3kRoJxox
gn9eg/lhjoWe9OHiYG5/KfHSPsT5TBA7qaBizVWPsIfThQoATXp1WPpEiWUaoB8a
+pra2GPKaJnGj+epxXu1p/S8lu+vhk54L/32BPvwwR0fSYe76ZFfMDlQcSUwrSH+
sUQuP/XmCTjTglzZ8pejtPmELWPZ0en94tLK57k/2+IR7prTl7C6K+ZGoSzluWw1
g61VEd/9AgMBAAGjUDBOMB0GA1UdDgQWBBQPOOpHijIhTlLnjvgELSvCeOxLhjAf
BgNVHSMEGDAWgBQPOOpHijIhTlLnjvgELSvCeOxLhjAMBgNVHRMEBTADAQH/MA0G
CSqGSIb3DQEBBQUAA4IBAQCjzghugH/Dmt0JCxQ/an9AW9KwYTuLhjWlsihdcuBa
HMnStX5JAY1CCu/pR2XCcOS7JnMC/MDGoxo4Kkt91S1xtB/TjabVqHWZActs0m6I
SDO79yLcK5xiwmU3VXJq2RAeFLBNXDrE5jos9TlfdO0ZAIMELsRTMcLWPN6YDNGd
o5cE9uqdqhU4ob6nkTNwBZF1AW8FB6Tm8vpeYQ8R4klLsXppMrF3eTEtJEAEjR+L
IYvk6ggJKPWFkXg8uIwGwDiHCZdt0KDW3aW8BR3Nb2TjcsPg8q1JLrgRSx2vb5yG
BYAngp+iVGvvOFORnucEhgvBmTCQNj2QgVWXHYcy6m5n
-----END CERTIFICATE-----`
	devicePrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQC2Nl4umpeykVae
f1XlH70gdpsAXT47kyTPrLzcLFQviDE2CNtOEjRUlj3Yu2yXkk5YLJz6AxnUy/Yr
UtzuldfzHJNs+lVtJ/BN3M4IZqsgW8VEfVO6H0R4qKng0SN0/dZWLwjx3kRoJxox
gn9eg/lhjoWe9OHiYG5/KfHSPsT5TBA7qaBizVWPsIfThQoATXp1WPpEiWUaoB8a
+pra2GPKaJnGj+epxXu1p/S8lu+vhk54L/32BPvwwR0fSYe76ZFfMDlQcSUwrSH+
sUQuP/XmCTjTglzZ8pejtPmELWPZ0en94tLK57k/2+IR7prTl7C6K+ZGoSzluWw1
g61VEd/9AgMBAAECggEBAJ8xvPNmqWafyFfN1JPMOQal2SJmVLrkAeNXfeG28Q5e
JPCaqE/7Y7ELWpmClouOjdhHkhZ0oxrh3/9V9CSW0gdbTgqGZCBsJm1AntmEKbqp
sx31iTMxojbw1QrQbfQG2j6N9XirwoRktWPQKcu/7h+nz2JjfP61VZTQZrBTmvBa
ZNLLO6w1AzrAaygWHkDMRUt3fsoTd2O9zOPgNITspdce2Tw2tc4+6qda2u0oSvC/
lewvkL0uUAYN/pWOwCwm02zH6VE56bNNOz2zBlif8/LkKQlDchpxE78JpH4ZbjtU
2XBO49rsLwVkUEIuo9TkZZgT/9fgFI9wovROLwMfmZkCgYEA3H2zQCAxtEaKw43Y
HfKChOrOGECywjFrXkrj+FkXpbrwZyFs/IINjqnvkmNxp4qUK7s10ZTck3htkeFA
mBePjY7kN5x80Gu4exdATpix40tC9RTuv/66Ij1aX9yfKUBLXJkOwWyvyT6D45p1
2cgzQjjFyQmDMbFJgKhWy3Z2jHsCgYEA046IrHdc0Kub+J3EJJS74e6+KftjGXMM
4xIuwiyeuUTAcEJvX1fOIkIgHmzEkJ1LKwKzPc3aFHMTyy3zhXFnmYNc8TxU/65U
CugzO4w+ulfseVqstO277oS5PocmcmBay7bcVFpvlW8OTgDew58zxrK3Oi87Bdtc
WnIaVKt5R+cCgYEAoZZpcFxnsMNl3IyuTrw0VO6znWiE2PZYxmDCE3ZPc0C+AAaq
FZ/GCcCWd0TzvSI9FpN7jJ24zUabniZjLVNO/CI1NGA1xJS9PVA7653R+E5mwq/V
jNVEWeV2vvwzlIqu8CyneK+LYEO1am7/YVxr3GM45+1VvWw8/tHf0fp+RNMCgYEA
mcjK4VQDTEzzHE7S/iSAT0RVR/9Nknpnq8jT5KK63sJzgSdJ/my9k3muD2/Rk65D
rghQc2ToWmUsxk2o8B/3x0gOj+3je9kljqgsVeUk1CCF7dFUKlGGg2RHpIRqFkqk
teE/WLJE2sPYCivnwxw/bvkK6Gjc5u0GvVike1gK2ZECgYB4IurXfRHRcQHVH/P3
IenX1suOOpgiqYqhzIzcH+MdxaVKmTmeCeCXJqVE1deMnal09kZQNlejn3QVncHZ
BY3cpXnM9ZhaIYy8/VqE9bbS07ieQA+kt8eKDeOZLNvseVBURa17XkUHIpl3FoiV
1oHTBYvtjxu6wdvTAgbomnIGcQ==
-----END PRIVATE KEY-----`
	mismatchedDevicePrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQC2Nl4umpeykVae
f1XlH70gdpsAXT47kyTPrLzcLFQviDE2CNtOEjRUlj3Yu2yXkk5YLJz6AxnUy/Yr
UtzuldfzHJNs+lVtJ/BN3M4IZqsgW8VEfVO6H0R4qKng0SN0/dZWLwjx3kRoJxox
gn9eg/lhjoWe9OHiYG5/KfHSPsT5TBA7qaBizVWPsIfThQoATXp1WPpEiWUaoB8a
+pra2GPKaJnGj+epxXu1p/S8lu+vhk54L/32BPvwwR0fSYe76ZFfMDlQcSUwrSH+
sUQuP/XmCTjTglzZ8pejtPmELWPZ0en94tLK57k/2+IR7prTl7C6K+ZGoSzluWw1
g61VEd/9AgMBAAECggEBAJ8xvPNmqWafyFfN1JPMOQal2SJmVLrkAeNXfeG28Q5e
JPCaqE/7Y7ELWpmClouOjdhHkhZ0oxrh3/9V9ABW0gdbTgqGZCBsJm1AntmEKbqp
sx31iTMxojbw1QrQbfQG2j6N9XirwoRktWPQKcu/7h+nz2JjfP61VZTQZrBTmvBa
ZNLLO6w1AzrAaygWHkDMRUt3fsoTd2O9zOPgNITspdce2Tw2tc4+6qda2u0oSvC/
lewvkL0uUAYN/pWOwCwm02zH6VE56bNNOz2zBlif8/LkKQlDchpxE78JpH4ZbjtU
2XBO49rsLwVkUEIuo9TkZZgT/9fgFI9wovROLwMfmZkCgYEA3H2zQCAxtEaKw43Y
HfKChOrOGECywjFrXkrj+FkXpbrwZyFs/IINjqnvkmNxp4qUK7s10ZTck3htkeFA
mBePjY7kN5x80Gu4exdATpix40tC9RTuv/66Ij1aX9yfKUBLXJkOwWyvyT6D45p1
2cgzQjjFyQmDMbFJgKhWy3Z2jHsCgYEA046IrHdc0Kub+J3EJJS74e6+KftjGXMM
4xIuwiyeuUTAcEJvX1fOIkIgHmzEkJ1LKwKzPc3aFHMTyy3zhXFnmYNc8TxU/65U
CugzO4w+ulfseVqstO277oS5PocmcmBay7bcVFpvlW8OTgDew58zxrK3Oi87Bdtc
WnIaVKt5R+cCgYEAoZZpcFxnsMNl3IyuTrw0VO6znWiE2PZYxmDCE3ZPc0C+AAaq
FZ/GCcCWd0TzvSI9FpN7jJ24zUabniZjLVNO/CI1NGA1xJS9PVA7653R+E5mwq/V
jNVEWeV2vvwzlIqu8CyneK+LYEO1am7/YVxr3GM45+1VvWw8/tHf0fp+RNMCgYEA
mcjK4VQDTEzzHE7S/iSAT0RVR/9Nknpnq8jT5KK63sJzgSdJ/my9k3muD2/Rk65D
rghQc2ToWmUsxk2o8B/3x0gOj+3je9kljqgsVeUk1CCF7dFUKlGGg2RHpIRqFkqk
teE/WLJE2sPYCivnwxw/bvkK6Gjc5u0GvVike1gK2ZECgYB4IurXfRHRcQHVH/P3
IenX1suOOpgiqYqhzIzcH+MdxaVKmTmeCeCXJqVE1deMnal09kZQNlejn3QVncHZ
BY3cpXnM9ZhaIYy8/VqE9bbS07ieQA+kt8eKDeOZLNvseVBURa17XkUHIpl3FoiV
1oHTBYvtjxu6wdvTAgbomnIGcQ==
-----END PRIVATE KEY-----`
	rootCACert = `
-----BEGIN CERTIFICATE-----
MIIDdzCCAl+gAwIBAgIEAgAAuTANBgkqhkiG9w0BAQUFADBaMQswCQYDVQQGEwJJ
RTESMBAGA1UEChMJQmFsdGltb3JlMRMwEQYDVQQLEwpDeWJlclRydXN0MSIwIAYD
VQQDExlCYWx0aW1vcmUgQ3liZXJUcnVzdCBSb290MB4XDTAwMDUxMjE4NDYwMFoX
DTI1MDUxMjIzNTkwMFowWjELMAkGA1UEBhMCSUUxEjAQBgNVBAoTCUJhbHRpbW9y
ZTETMBEGA1UECxMKQ3liZXJUcnVzdDEiMCAGA1UEAxMZQmFsdGltb3JlIEN5YmVy
VHJ1c3QgUm9vdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKMEuyKr
mD1X6CZymrV51Cni4eiVgLGw41uOKymaZN+hXe2wCQVt2yguzmKiYv60iNoS6zjr
IZ3AQSsBUnuId9Mcj8e6uYi1agnnc+gRQKfRzMpijS3ljwumUNKoUMMo6vWrJYeK
mpYcqWe4PwzV9/lSEy/CG9VwcPCPwBLKBsua4dnKM3p31vjsufFoREJIE9LAwqSu
XmD+tqYF/LTdB1kC1FkYmGP1pWPgkAx9XbIGevOF6uvUA65ehD5f/xXtabz5OTZy
dc93Uk3zyZAsuT3lySNTPx8kmCFcB5kpvcY67Oduhjprl3RjM71oGDHweI12v/ye
jl0qhqdNkNwnGjkCAwEAAaNFMEMwHQYDVR0OBBYEFOWdWTCCR1jMrPoIVDaGezq1
BE3wMBIGA1UdEwEB/wQIMAYBAf8CAQMwDgYDVR0PAQH/BAQDAgEGMA0GCSqGSIb3
DQEBBQUAA4IBAQCFDF2O5G9RaEIFoN27TyclhAO992T9Ldcw46QQF+vaKSm2eT92
9hkTI7gQCvlYpNRhcL0EYWoSihfVCr3FvDB81ukMJY2GQE/szKN+OMY3EU/t3Wgx
jkzSswF07r51XgdIGn9w/xZchMB5hbgF/X++ZRGjD8ACtPhSNzkE1akxehi/oCr0
Epn3o0WC4zxe9Z2etciefC7IpJ5OCBRLbf1wbWsaY71k5h+3zvDyny67G7fyUIhz
ksLi4xaNmjICq44Y3ekQEe5+NauQrz4wlHrQMz2nZQ/1/I6eYs9HRCwBXbsdtTLS
R9I4LtD+gdwyah617jzV/OeBHRnDJELqYzmp
-----END CERTIFICATE-----`
)

func TestRegularMQTTSendNoParams(t *testing.T) {
	sender := NewRegularMQTTSender(RegularMQTTConfig{}, true, false)
	continuePipeline, result := sender.Send(ctx, nil)
	assert.False(t, continuePipeline)
	assert.Error(t, result.(error))
}

func TestRegularMQTTSendSetRetryDataPersistFalse(t *testing.T) {
	sender := NewRegularMQTTSender(RegularMQTTConfig{}, false, false)
	ctx.SetRetryData(nil)
	sender.setRetryData(ctx, []byte("data"))
	assert.Nil(t, ctx.RetryData())
}

func TestRegularMQTTSendSetRetryDataPersistTrue(t *testing.T) {
	sender := NewRegularMQTTSender(RegularMQTTConfig{}, true, false)
	ctx.SetRetryData(nil)
	sender.setRetryData(ctx, []byte("data"))
	assert.Equal(t, []byte("data"), ctx.RetryData())
}

func TestConfigureMQTTClientForAuthUsernamePassword(t *testing.T) {
	factory := RegularMqttClientFactory{
		authMode: messaging.AuthModeUsernamePassword,
		opts:     MQTT.NewClientOptions(),
	}
	secrets := mqttSecrets{
		username: "iotech",
		password: "awesome",
	}
	err := factory.configureMQTTClientForAuth(secrets)
	require.NoError(t, err)
	assert.Nil(t, factory.opts.TLSConfig.Certificates)
	assert.Equal(t, factory.opts.Username, "iotech")
	assert.Equal(t, factory.opts.Password, "awesome")
}

func TestConfigureMQTTClientForAuthUsernamePasswordAndCA(t *testing.T) {
	factory := RegularMqttClientFactory{
		authMode: messaging.AuthModeUsernamePassword,
		opts:     MQTT.NewClientOptions(),
	}
	secrets := mqttSecrets{
		username:   "iotech",
		password:   "awesome",
		caPemBlock: []byte(rootCACert),
	}
	err := factory.configureMQTTClientForAuth(secrets)
	require.NoError(t, err)
	assert.Nil(t, factory.opts.TLSConfig.Certificates)
	assert.Equal(t, factory.opts.Username, "iotech")
	assert.Equal(t, factory.opts.Password, "awesome")
	assert.NotNil(t, factory.opts.TLSConfig.RootCAs)
}

func TestConfigureMQTTClientForAuthWithCACert(t *testing.T) {
	username := "iotech"
	factory := RegularMqttClientFactory{
		authMode: messaging.AuthModeCA,
		opts:     MQTT.NewClientOptions(),
	}
	secrets := mqttSecrets{
		username:   username,
		password:   "awesome",
		caPemBlock: []byte(rootCACert),
	}
	err := factory.configureMQTTClientForAuth(secrets)
	require.NoError(t, err)
	assert.NotNil(t, factory.opts.TLSConfig.RootCAs)
	assert.Equal(t, factory.opts.Username, username)
	assert.Empty(t, factory.opts.Password)
	assert.Nil(t, factory.opts.TLSConfig.Certificates)
}

func TestConfigureMQTTClientForAuthWithClientCert(t *testing.T) {
	username := "iotech"
	factory := RegularMqttClientFactory{
		authMode: messaging.AuthModeCert,
		opts:     MQTT.NewClientOptions(),
	}
	secrets := mqttSecrets{
		username:     username,
		password:     "awesome",
		certPemBlock: []byte(deviceCert),
		keyPemBlock:  []byte(devicePrivateKey),
		caPemBlock:   []byte(rootCACert),
	}
	err := factory.configureMQTTClientForAuth(secrets)
	require.NoError(t, err)
	assert.Equal(t, factory.opts.Username, username)
	assert.Empty(t, factory.opts.Password)
	assert.NotNil(t, factory.opts.TLSConfig.Certificates)
	assert.NotNil(t, factory.opts.TLSConfig.RootCAs)
}

func TestConfigureMQTTClientForAuthWithClientCertNoCA(t *testing.T) {
	username := "iotech"
	factory := RegularMqttClientFactory{
		authMode: messaging.AuthModeCert,
		opts:     MQTT.NewClientOptions(),
	}
	secrets := mqttSecrets{
		username:     username,
		password:     "awesome",
		certPemBlock: []byte(deviceCert),
		keyPemBlock:  []byte(devicePrivateKey),
	}
	err := factory.configureMQTTClientForAuth(secrets)
	require.NoError(t, err)
	assert.Equal(t, factory.opts.Username, username)
	assert.Empty(t, factory.opts.Password)
	assert.NotNil(t, factory.opts.TLSConfig.Certificates)
	assert.Nil(t, factory.opts.TLSConfig.RootCAs)
}

func TestConfigureMQTTClientForAuthWithNone(t *testing.T) {
	factory := RegularMqttClientFactory{
		authMode: messaging.AuthModeNone,
		opts:     MQTT.NewClientOptions(),
	}
	err := factory.configureMQTTClientForAuth(mqttSecrets{})
	require.NoError(t, err)
}
