// Copyright (C) 2021 IOTech Ltd

package xpert

import (
	"errors"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	watsonUsername = "use-token-auth"
	watsonPassword = "K*&WB2!9nvY*NG-kzF" //nolint: gosec
	watsonCACert   = `-----BEGIN CERTIFICATE-----
MIIElDCCA3ygAwIBAgIQAf2j627KdciIQ4tyS8+8kTANBgkqhkiG9w0BAQsFADBh
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD
QTAeFw0xMzAzMDgxMjAwMDBaFw0yMzAzMDgxMjAwMDBaME0xCzAJBgNVBAYTAlVT
MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxJzAlBgNVBAMTHkRpZ2lDZXJ0IFNIQTIg
U2VjdXJlIFNlcnZlciBDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
ANyuWJBNwcQwFZA1W248ghX1LFy949v/cUP6ZCWA1O4Yok3wZtAKc24RmDYXZK83
nf36QYSvx6+M/hpzTc8zl5CilodTgyu5pnVILR1WN3vaMTIa16yrBvSqXUu3R0bd
KpPDkC55gIDvEwRqFDu1m5K+wgdlTvza/P96rtxcflUxDOg5B6TXvi/TC2rSsd9f
/ld0Uzs1gN2ujkSYs58O09rg1/RrKatEp0tYhG2SS4HD2nOLEpdIkARFdRrdNzGX
kujNVA075ME/OV4uuPNcfhCOhkEAjUVmR7ChZc6gqikJTvOX6+guqw9ypzAO+sf0
/RR3w6RbKFfCs/mC/bdFWJsCAwEAAaOCAVowggFWMBIGA1UdEwEB/wQIMAYBAf8C
AQAwDgYDVR0PAQH/BAQDAgGGMDQGCCsGAQUFBwEBBCgwJjAkBggrBgEFBQcwAYYY
aHR0cDovL29jc3AuZGlnaWNlcnQuY29tMHsGA1UdHwR0MHIwN6A1oDOGMWh0dHA6
Ly9jcmwzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydEdsb2JhbFJvb3RDQS5jcmwwN6A1
oDOGMWh0dHA6Ly9jcmw0LmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydEdsb2JhbFJvb3RD
QS5jcmwwPQYDVR0gBDYwNDAyBgRVHSAAMCowKAYIKwYBBQUHAgEWHGh0dHBzOi8v
d3d3LmRpZ2ljZXJ0LmNvbS9DUFMwHQYDVR0OBBYEFA+AYRyCMWHVLyjnjUY4tCzh
xtniMB8GA1UdIwQYMBaAFAPeUDVW0Uy7ZvCj4hsbw5eyPdFVMA0GCSqGSIb3DQEB
CwUAA4IBAQAjPt9L0jFCpbZ+QlwaRMxp0Wi0XUvgBCFsS+JtzLHgl4+mUwnNqipl
5TlPHoOlblyYoiQm5vuh7ZPHLgLGTUq/sELfeNqzqPlt/yGFUzZgTHbO7Djc1lGA
8MXW5dRNJ2Srm8c+cftIl7gzbckTB+6WohsYFfZcTEDts8Ls/3HB40f/1LkAtDdC
2iDJ6m6K7hQGrn2iWZiIqBtvLfTyyRRfJs8sjX7tN8Cp1Tm5gr8ZDOo0rwAhaPit
c+LJMto4JQtV05od8GiG7S5BNO98pVAdvzr508EIDObtHopYJeS4d60tbvVS3bR0
j6tJLp07kzQoH3jOlOrHvdPJbRzeXDLz
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDrzCCApegAwIBAgIQCDvgVpBCRrGhdWrJWZHHSjANBgkqhkiG9w0BAQUFADBh
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD
QTAeFw0wNjExMTAwMDAwMDBaFw0zMTExMTAwMDAwMDBaMGExCzAJBgNVBAYTAlVT
MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5j
b20xIDAeBgNVBAMTF0RpZ2lDZXJ0IEdsb2JhbCBSb290IENBMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4jvhEXLeqKTTo1eqUKKPC3eQyaKl7hLOllsB
CSDMAZOnTjC3U/dDxGkAV53ijSLdhwZAAIEJzs4bg7/fzTtxRuLWZscFs3YnFo97
nh6Vfe63SKMI2tavegw5BmV/Sl0fvBf4q77uKNd0f3p4mVmFaG5cIzJLv07A6Fpt
43C/dxC//AH2hdmoRBBYMql1GNXRor5H4idq9Joz+EkIYIvUX7Q6hL+hqkpMfT7P
T19sdl6gSzeRntwi5m3OFBqOasv+zbMUZBfHWymeMr/y7vrTC0LUq7dBMtoM1O/4
gdW7jVg/tRvoSSiicNoxBN33shbyTApOB6jtSj1etX+jkMOvJwIDAQABo2MwYTAO
BgNVHQ8BAf8EBAMCAYYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUA95QNVbR
TLtm8KPiGxvDl7I90VUwHwYDVR0jBBgwFoAUA95QNVbRTLtm8KPiGxvDl7I90VUw
DQYJKoZIhvcNAQEFBQADggEBAMucN6pIExIK+t1EnE9SsPTfrgT1eXkIoyQY/Esr
hMAtudXH/vTBH1jLuG2cenTnmCmrEbXjcKChzUyImZOMkXDiqw8cvpOp/2PV5Adg
06O/nVsJ8dWO41P0jmP6P6fbtGbfYmbW0W5BjfIttep3Sp+dWOIrWcBAI+0tKIJF
PnlUkiaY4IBIqDfv8NZ5YBberOgOzW6sRBc4L0na4UU+Krk2U886UAb3LujEV0ls
YSEY1QSteDwsOoBrp+uvFRTp2InBuThs4pFsiv9kuXclVzDAGySj4dzp30d8tbQk
CAUw7C29C79Fv1C5qfPrmAESrciIxpg0X40KPMbp1ZWVbd4=
-----END CERTIFICATE-----
`
	invalidWatsonCACert = "invalid ca cert"
)

func TestWatsonSendNoParams(t *testing.T) {
	sender := NewIBMWatsonMQTTSender(IBMWatsonMQTTConfig{}, true, false)
	continuePipeline, result := sender.Send(ctx, nil)
	require.False(t, continuePipeline)
	require.Error(t, result.(error))
}

func TestWatsonSendSetRetryDataPersistFalse(t *testing.T) {
	sender := NewIBMWatsonMQTTSender(IBMWatsonMQTTConfig{}, false, false)
	ctx.SetRetryData(nil)
	sender.setRetryData(ctx, []byte("data"))
	assert.Nil(t, ctx.RetryData())
}

func TestWatsonSendSetRetryDataPersistTrue(t *testing.T) {
	sender := NewIBMWatsonMQTTSender(IBMWatsonMQTTConfig{}, true, false)
	ctx.SetRetryData(nil)
	sender.setRetryData(ctx, []byte("data"))
	assert.Equal(t, []byte("data"), ctx.RetryData())
}

func TestWatsonSendValidateSecrets(t *testing.T) {
	factory := IBMWatsonMqttClientFactory{opts: MQTT.NewClientOptions()}
	tests := []struct {
		Name             string
		secrets          watsonSecrets
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{"No username", watsonSecrets{password: watsonPassword, certpemblock: []byte(watsonCACert)}, true, "mandatory username is empty"},
		{"No password", watsonSecrets{username: watsonUsername, certpemblock: []byte(watsonCACert)}, true, "mandatory password is empty"},
		{"All satisfied", watsonSecrets{username: watsonUsername, password: watsonPassword}, false, ""},
		{"Valid CA certificate", watsonSecrets{username: watsonUsername, password: watsonPassword, certpemblock: []byte(watsonCACert)}, false, ""},
		{"Invalid CA certificate", watsonSecrets{username: watsonUsername, password: watsonPassword, certpemblock: []byte(invalidWatsonCACert)}, true, "error parsing CA certificate PEM block"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := factory.configureMQTTClientForAuth(test.secrets)
			if test.ErrorExpectation {
				assert.Error(t, result, "Result should be an error")
				assert.Equal(t, test.ErrorMessage, result.Error())
			} else {
				assert.Nil(t, result, "Should be nil")
				assert.Equal(t, watsonUsername, factory.opts.Username)
				assert.Equal(t, watsonPassword, factory.opts.Password)
				assert.NotNil(t, factory.opts.TLSConfig.ClientCAs)
			}
		})
	}
}

func TestWatsonSendGetSecrets(t *testing.T) {
	factory := IBMWatsonMqttClientFactory{opts: MQTT.NewClientOptions()}
	mockSP := &mocks.SecretProvider{}
	mockSP.On("GetSecret", "/notfound").Return(nil, errors.New(""))
	mockSP.On("GetSecret", "/watson").Return(
		map[string]string{
			MQTTSecretUsername: watsonUsername,
			MQTTSecretPassword: watsonPassword,
			MQTTSecretCACert:   watsonCACert,
		}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})
	tests := []struct {
		Name            string
		SecretPath      string
		ExpectedSecrets *watsonSecrets
		ExpectingError  bool
	}{
		{"No Secrets found", "/notfound", nil, true},
		{"With Secrets", "/watson", &watsonSecrets{
			username:     watsonUsername,
			password:     watsonPassword,
			certpemblock: []byte(watsonCACert),
		}, false},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			factory.secretPath = test.SecretPath
			watsonSecrets, err := factory.getSecrets(ctx)
			if test.ExpectingError {
				assert.Error(t, err, "Expecting error")
				return
			}
			require.Equal(t, test.ExpectedSecrets, watsonSecrets)
		})
	}
}
