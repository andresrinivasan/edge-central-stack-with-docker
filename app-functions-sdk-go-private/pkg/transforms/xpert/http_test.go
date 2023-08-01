// Copyright (C) 2022 IOTech Ltd

package xpert

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	mocks2 "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	msgStr         = "test message"
	urlSubPath1    = "foo"
	urlSubPath2    = "bad"
	baseUrlPath    = "/base/"
	urlPath        = baseUrlPath + urlSubPath1
	badPath        = baseUrlPath + urlSubPath2
	urlParamKey    = "test"
	urlFormatParam = "{" + urlParamKey + "}"
	formatUrl      = baseUrlPath + urlFormatParam
	badFormatUrl   = baseUrlPath + urlFormatParam + "/{test2}"
)

func TestXpertHTTPExportNoAuth(t *testing.T) {
	var methodReceived string

	handler := func(w http.ResponseWriter, r *http.Request) {
		methodReceived = r.Method

		if r.URL.EscapedPath() == badPath {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)

		readMsg, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		if strings.Compare((string)(readMsg), msgStr) != 0 {
			t.Errorf("Invalid msg received %v, expected %v", readMsg, msgStr)
		}

		if r.Header.Get(common.ContentType) != common.ContentTypeJSON {
			t.Errorf("Unexpected content-type received %s, expected %s", r.Header.Get(common.ContentType), common.ContentTypeJSON)
		}
		if r.URL.EscapedPath() != urlPath {
			t.Errorf("Invalid urlPath received %s, expected %s",
				r.URL.EscapedPath(), urlPath)
		}
	}

	mockSP := &mocks2.SecretProvider{}
	mockSP.On("GetSecret", "").Return(map[string]string{}, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	targetUrl, err := url.Parse(ts.URL)
	require.NoError(t, err)

	tests := []struct {
		Name                      string
		Path                      string
		HTTPMethod                string
		PersistOnError            bool
		RetryDataSet              bool
		ReturnInputData           bool
		ContinueOnSendError       bool
		AuthMode                  HTTPAuthMode
		ExpectedContinueExecuting bool
	}{
		{"Successful GET", urlPath, http.MethodGet, true, false, false, false, HTTPAuthModeNONE, true},
		{"Successful GET with Format URL", formatUrl, http.MethodGet, true, false, false, false, HTTPAuthModeNONE, true},
		{"Successful POST", urlPath, http.MethodPost, true, false, false, false, HTTPAuthModeNONE, true},
		{"Successful POST with Format URL", formatUrl, http.MethodPost, true, false, false, false, HTTPAuthModeNONE, true},
		{"Successful PUT", urlPath, http.MethodPut, true, false, false, false, HTTPAuthModeNONE, true},
		{"Successful PUT with Format URL", formatUrl, http.MethodPut, true, false, false, false, HTTPAuthModeNONE, true},
		{"Successful PATCH", urlPath, http.MethodPatch, true, false, false, false, HTTPAuthModeNONE, true},
		{"Successful PATCH with Format URL", formatUrl, http.MethodPatch, true, false, false, false, HTTPAuthModeNONE, true},
		{"Successful DELETE", urlPath, http.MethodDelete, true, false, false, false, HTTPAuthModeNONE, true},
		{"Successful DELETE with Format URL", formatUrl, http.MethodDelete, true, false, false, false, HTTPAuthModeNONE, true},
		{"Failed case - bad url with no persist", badPath, http.MethodPost, false, false, false, false, HTTPAuthModeNONE, false},
		{"Failed case - bad url with continueOnError and returnInputData", badPath, http.MethodPost, false, false, true, true, HTTPAuthModeNONE, true},
		{"Failed case - bad url with persistOnError", badPath, http.MethodPost, true, true, false, false, HTTPAuthModeNONE, false},
		{"Successful return input data", urlPath, http.MethodPost, false, false, true, false, HTTPAuthModeNONE, true},
		{"Failed case - persist and return input data", badPath, http.MethodPost, true, true, true, false, HTTPAuthModeNONE, false},
		{"Successful continueOnSendError and ignore returnInputData", urlPath, http.MethodPost, false, false, false, true, HTTPAuthModeNONE, true},
		{"Successful continueOnSendError and ignore persistOnError", badPath, http.MethodPost, true, false, true, true, HTTPAuthModeNONE, true},
		{"Failed case - bad format URL", badFormatUrl, "", true, false, false, false, HTTPAuthModeNONE, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx.AddValue(urlParamKey, urlSubPath1)
			methodReceived = ""
			ctx.SetRetryData(nil)
			config := HTTPSenderConfig{
				HTTPSenderOptions: transforms.HTTPSenderOptions{
					URL:                 "http://" + targetUrl.Host + test.Path,
					MimeType:            common.ContentTypeJSON,
					PersistOnError:      test.PersistOnError,
					ContinueOnSendError: test.ContinueOnSendError,
					ReturnInputData:     test.ReturnInputData,
				},
				HTTPMethod: test.HTTPMethod,
				AuthMode:   test.AuthMode,
			}
			sender := NewHTTPSender(config)
			var continueExecuting bool
			var resultData interface{}

			continueExecuting, resultData = sender.HTTPSend(ctx, msgStr)

			assert.Equal(t, test.ExpectedContinueExecuting, continueExecuting)

			if test.ExpectedContinueExecuting {
				if test.ReturnInputData {
					assert.Equal(t, msgStr, resultData)
				} else {
					assert.NotEqual(t, msgStr, resultData)
				}
			}
			assert.Equal(t, test.RetryDataSet, ctx.RetryData() != nil)
			assert.Equal(t, test.HTTPMethod, methodReceived)
			ctx.RemoveValue(urlParamKey)
		})
	}
}

func TestXpertHTTPExportHeaderSecrets(t *testing.T) {
	var methodUsed string

	expectedSecretValue := "my-API-key"
	headerName := "Secret-Header-Name"
	secretPath := "secretPath"
	validSecretName := "header"
	invalidSecretName := "bogus"

	mockSP := &mocks2.SecretProvider{}
	mockSP.On("GetSecret", secretPath).Return(map[string]string{validSecretName: expectedSecretValue}, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})

	placeholderCheck := regexp.MustCompile("{[^}]*}")

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		methodUsed = request.Method

		if request.URL.EscapedPath() == badPath {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		if placeholderCheck.MatchString(request.URL.RawPath) {
			writer.WriteHeader(http.StatusBadRequest)
			require.Fail(t, "url placeholders not replaced")
		}

		writer.WriteHeader(http.StatusOK)

		actualValue := request.Header.Get(headerName)

		if actualValue != "" {
			// Only validate is key was found in the header
			require.Equal(t, expectedSecretValue, actualValue)
		}
	}))
	defer ts.Close()

	targetUrl, err := url.Parse(ts.URL)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		Path                 string
		HeaderName           string
		SecretName           string
		SecretPath           string
		ExpectToContinue     bool
		ExpectedErrorMessage string
		ExpectedMethod       string
	}{
		{"unsuccessful w/o secret header name", urlPath, "", validSecretName, secretPath, false, "HTTP Header Name is required", ""},
		{"unsuccessful w/o secret urlPath", urlPath, headerName, validSecretName, "", false, "secretPath is required", ""},
		{"unsuccessful w/o secret name", urlPath, headerName, "", secretPath, false, "secretName is required", ""},
		{"successful with secrets", urlPath, headerName, validSecretName, secretPath, true, "", http.MethodPost},
		{"successful with secrets and formatted urlPath", formatUrl, headerName, validSecretName, secretPath, true, "", http.MethodPost},
		{"unsuccessful with secrets - retrieval fails", urlPath, headerName, invalidSecretName, secretPath, false, "no corresponding secret can be found", ""},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx.AddValue(urlParamKey, urlSubPath1)
			methodUsed = ""
			config := HTTPSenderConfig{
				HTTPSenderOptions: transforms.HTTPSenderOptions{
					URL:            "http://" + targetUrl.Host + test.Path,
					MimeType:       common.ContentTypeJSON,
					PersistOnError: false,
					HTTPHeaderName: test.HeaderName,
					SecretPath:     test.SecretPath,
					SecretName:     test.SecretName,
				},
				HTTPMethod: http.MethodPost,
				AuthMode:   HTTPAuthModeHeaderSecret,
			}
			sender := NewHTTPSender(config)

			var continuePipeline bool
			var err interface{}

			continuePipeline, err = sender.HTTPSend(ctx, msgStr)

			assert.Equal(t, test.ExpectToContinue, continuePipeline)
			if !test.ExpectToContinue {
				require.Contains(t, err.(error).Error(), test.ExpectedErrorMessage)
			}
			assert.Equal(t, test.ExpectedMethod, methodUsed)
			ctx.RemoveValue(urlParamKey)
		})
	}
}

func retrieveTokenValue(request *http.Request, headerKey string, separator string, index int) string {
	reqToken := request.Header.Get(headerKey)
	splitToken := strings.Split(reqToken, separator)
	if len(splitToken) >= index {
		return splitToken[index]
	}
	return ""
}

func TestXpertHTTPExportOAuth2ClientCredentials(t *testing.T) {
	var methodUsed string

	secretPath := "mySecret"
	secretPathWithInsufficientSecrets := "insufficient"
	invalidSecretPath := "bogus"
	clientId := "hF2FNwmhOrc7uA32k13dnhmY5lFLp9O9"
	clientSecret := "ZO-u2SinnhkLzqFwUV-n3s16ndI2WKqw09SKNa33v5qgvKFkptYeynjnBQCugpCR" //nolint: gosec
	tokenScopes := "scope1"
	tokenPath := "/token"
	expectedTokenValue := "abc123"
	tokenTypeBasic := "Basic"
	tokenTypeBearer := "Bearer"
	headerKeyAuthorization := "Authorization"

	placeholderCheck := regexp.MustCompile("{[^}]*}")

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		methodUsed = request.Method

		switch request.URL.EscapedPath() {
		case badPath:
			writer.WriteHeader(http.StatusNotFound)
			return
		case tokenPath: //oauth2 token endpoint simulation
			auth := clientId + ":" + clientSecret
			encodedStr := base64.StdEncoding.EncodeToString([]byte(auth))
			reqToken := retrieveTokenValue(request, headerKeyAuthorization, tokenTypeBasic+" ", 1)
			require.True(t, len(reqToken) > 0, "fail to retrieve token from header")
			require.Equal(t, encodedStr, reqToken, "unexpected token value")
			writer.Header().Set(common.ContentType, common.ContentTypeJSON)
			writer.WriteHeader(http.StatusOK)
			jsonTokenResponse := map[string]interface{}{
				"access_token": expectedTokenValue,
				"scope":        tokenScopes,
				"token_type":   tokenTypeBearer,
			}
			err := json.NewEncoder(writer).Encode(jsonTokenResponse)
			require.NoError(t, err)
			return
		}

		if placeholderCheck.MatchString(request.URL.RawPath) {
			writer.WriteHeader(http.StatusBadRequest)
			require.Fail(t, "url placeholders not replaced")
		}

		writer.WriteHeader(http.StatusOK)
		reqToken := retrieveTokenValue(request, headerKeyAuthorization, tokenTypeBearer+" ", 1)
		require.True(t, len(reqToken) > 0, "fail to retrieve token from header")
		require.Equal(t, expectedTokenValue, reqToken, "unexpected token value")
	}))
	defer ts.Close()

	targetUrl, err := url.Parse(ts.URL)
	require.NoError(t, err)

	mockSP := &mocks2.SecretProvider{}
	mockSP.On("GetSecret", secretPath).Return(map[string]string{SecretKeyClientId: clientId, SecretKeyClientSecret: clientSecret, SecretKeyTokenScopes: tokenScopes, SecretKeyTokenUrl: "http://" + targetUrl.Host + tokenPath}, nil)
	mockSP.On("GetSecret", secretPathWithInsufficientSecrets).Return(map[string]string{SecretKeyClientId: clientId, SecretKeyTokenScopes: tokenScopes, SecretKeyTokenUrl: "http://" + targetUrl.Host + tokenPath}, nil)
	mockSP.On("GetSecret", invalidSecretPath).Return(nil, errors.New("FAKE NOT FOUND ERROR"))

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})

	tests := []struct {
		Name                 string
		Path                 string
		SecretPath           string
		ExpectToContinue     bool
		ExpectedErrorMessage string
		ExpectedMethod       string
	}{
		{"successful case", urlPath, secretPath, true, "", http.MethodPost},
		{"unsuccessful - invalid secret path", urlPath, invalidSecretPath, false, "fail to retrieve secrets", ""},
		{"unsuccessful - secret path with insufficient secrets", urlPath, secretPathWithInsufficientSecrets, false, "doesn't exist or is empty", ""},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx.AddValue(urlParamKey, urlSubPath1)
			methodUsed = ""
			config := HTTPSenderConfig{
				HTTPSenderOptions: transforms.HTTPSenderOptions{
					URL:            "http://" + targetUrl.Host + test.Path,
					MimeType:       common.ContentTypeJSON,
					PersistOnError: false,
					SecretPath:     test.SecretPath,
				},
				HTTPMethod: http.MethodPost,
				AuthMode:   HTTPAuthModeOauth2ClientCredentials,
			}
			sender := NewHTTPSender(config)

			var continuePipeline bool
			var err interface{}

			continuePipeline, err = sender.HTTPSend(ctx, msgStr)

			assert.Equal(t, test.ExpectToContinue, continuePipeline)
			if !test.ExpectToContinue {
				require.Contains(t, err.(error).Error(), test.ExpectedErrorMessage)
			}
			assert.Equal(t, test.ExpectedMethod, methodUsed)
			ctx.RemoveValue(urlParamKey)
		})
	}
}

func TestXpertHTTPExportClientCert(t *testing.T) {
	var methodUsed string

	secretPath := "secretPath"
	secretPathNotFound := "notFound"
	secretPathMismatchedKeyCert := "mismatchedKeyCert"

	mockSP := &mocks2.SecretProvider{}
	mockSP.On("GetSecret", secretPath).Return(map[string]string{
		messaging.SecretClientCert: deviceCert,
		messaging.SecretClientKey:  devicePrivateKey,
	}, nil)
	mockSP.On("GetSecret", secretPathNotFound).Return(map[string]string{}, nil)
	mockSP.On("GetSecret", secretPathMismatchedKeyCert).Return(map[string]string{
		messaging.SecretClientCert: deviceCert,
		messaging.SecretClientKey:  mismatchedDevicePrivateKey,
	}, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})

	// create test server with handler
	ts := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		methodUsed = request.Method
		writer.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	targetUrl, err := url.Parse(ts.URL)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		SecretPath           string
		ExpectToContinue     bool
		ExpectedErrorMessage string
		ExpectedMethod       string
	}{
		{"successful with secrets", secretPath, true, "", http.MethodPost},
		{"unsuccessful w/o secret path", "", false, "secretPath is required", ""},
		{"unsuccessful with secrets - retrieval fails", secretPathNotFound, false, "no corresponding secret can be found", ""},
		{"unsuccessful with secrets - mismatched key/cert", secretPathMismatchedKeyCert, false, "error parsing client cert", ""},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			methodUsed = ""
			config := HTTPSenderConfig{
				HTTPSenderOptions: transforms.HTTPSenderOptions{
					URL:            "https://" + targetUrl.Host,
					MimeType:       common.ContentTypeJSON,
					PersistOnError: false,
					SecretPath:     test.SecretPath,
				},
				HTTPMethod:     http.MethodPost,
				AuthMode:       HTTPAuthModeClientCert,
				SkipCertVerify: true,
			}
			sender := NewHTTPSender(config)

			var continuePipeline bool
			var err interface{}

			continuePipeline, err = sender.HTTPSend(ctx, msgStr)

			assert.Equal(t, test.ExpectToContinue, continuePipeline)
			if !test.ExpectToContinue {
				require.Contains(t, err.(error).Error(), test.ExpectedErrorMessage)
			}
			assert.Equal(t, test.ExpectedMethod, methodUsed)
		})
	}
}

func TestXpertHTTPExportAWSSignature(t *testing.T) {
	var methodUsed string

	secretPath := "secretPath"
	secretPathMissingAWSAccessKey := "missingAWSAccessKey"
	secretPathMissingAWSSecretKey := "missingAWSSecretKey"
	validAWSAccessKey := "access"
	validAWSSecretKey := "secret"

	mockSP := &mocks2.SecretProvider{}
	mockSP.On("GetSecret", secretPath).Return(map[string]string{
		SecretKeyAWSAccessKey: validAWSAccessKey,
		SecretKeyAWSSecretKey: validAWSSecretKey,
	}, nil)
	mockSP.On("GetSecret", secretPathMissingAWSAccessKey).Return(map[string]string{
		SecretKeyAWSSecretKey: validAWSSecretKey,
	}, nil)
	mockSP.On("GetSecret", secretPathMissingAWSSecretKey).Return(map[string]string{
		SecretKeyAWSAccessKey: validAWSAccessKey,
	}, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		methodUsed = request.Method
		writer.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	targetUrl, err := url.Parse(ts.URL)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		Path                 string
		SecretPath           string
		ExpectToContinue     bool
		ExpectedErrorMessage string
		ExpectedMethod       string
	}{
		{"successful with secrets", urlPath, secretPath, true, "", http.MethodPost},
		{"unsuccessful with secrets - no access key", urlPath, secretPathMissingAWSAccessKey, false, "secret aws_access_key doesn't exist or is empty", ""},
		{"unsuccessful with secrets - no secret key", urlPath, secretPathMissingAWSSecretKey, false, "secret aws_secret_key doesn't exist or is empty", ""},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx.AddValue(urlParamKey, urlSubPath1)
			methodUsed = ""
			config := HTTPSenderConfig{
				HTTPSenderOptions: transforms.HTTPSenderOptions{
					URL:            "http://" + targetUrl.Host + test.Path,
					MimeType:       common.ContentTypeJSON,
					PersistOnError: false,
					SecretPath:     test.SecretPath,
				},
				HTTPMethod:         http.MethodPost,
				AuthMode:           HTTPAuthModeAWSSignature,
				AWSV4SignerConfigs: map[string]string{"region": "us-east-1", "service": "s3"},
			}
			sender := NewHTTPSender(config)

			var continuePipeline bool
			var err interface{}

			continuePipeline, err = sender.HTTPSend(ctx, msgStr)
			assert.Equal(t, test.ExpectToContinue, continuePipeline)
			if !test.ExpectToContinue {
				require.Contains(t, err.(error).Error(), test.ExpectedErrorMessage)
			}
			assert.Equal(t, test.ExpectedMethod, methodUsed)
			ctx.RemoveValue(urlParamKey)
		})
	}
}
