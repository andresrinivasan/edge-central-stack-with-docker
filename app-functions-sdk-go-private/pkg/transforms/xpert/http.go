// Copyright (C) 2022-2023 IOTech Ltd

package xpert

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	gometrics "github.com/rcrowley/go-metrics"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsv4signer "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

type HTTPAuthMode string

const (
	HTTPAuthModeNONE                    HTTPAuthMode = "none"
	HTTPAuthModeHeaderSecret            HTTPAuthMode = "header_secret"
	HTTPAuthModeOauth2ClientCredentials HTTPAuthMode = "oauth2_clientcredentials" //nolint: gosec
	HTTPAuthModeClientCert              HTTPAuthMode = messaging.AuthModeCert
	HTTPAuthModeAWSSignature            HTTPAuthMode = "aws_signature"

	SecretKeyClientId        = "client_id"
	SecretKeyClientSecret    = "client_secret"
	SecretKeyTokenScopes     = "token_scopes"
	SecretKeyTokenUrl        = "token_url"
	SecretKeyTokenServerCert = "token_cert"
	SecretKeyAWSAccessKey    = "aws_access_key"
	SecretKeyAWSSecretKey    = "aws_secret_key"

	AWSV4SignerConfigsRegion  = "region"
	AWSV4SignerConfigsService = "service"
)

type HTTPSenderConfig struct {
	transforms.HTTPSenderOptions
	// HTTPMethod is the method of http request
	HTTPMethod string
	// AuthMode specifies what HTTP authentication mode to use, now supports "none", "header_secret", and
	// "oauth2_clientcredentials".  The default value will be none.
	AuthMode HTTPAuthMode
	// SkipCertVerify indicates whether to verify the server's certificate chain and host name
	SkipCertVerify bool
	// Renegotiation controls what types of renegotiation are supported.
	// 0: disables renegotiation
	// 1: allows a remote server to request renegotiation once per connection
	// 2: allows a remote server to repeatedly request renegotiation
	RenegotiationSupport tls.RenegotiationSupport
	HTTPRequestHeaders   map[string]string
	AWSV4SignerConfigs   map[string]string
}

// NewHTTPSender - factory method for creation of HTTPSender
func NewHTTPSender(senderConfig HTTPSenderConfig) *HTTPSender {
	return &HTTPSender{
		config: senderConfig,
	}
}

// HTTPSender is the proprietary implementation to support wider range of http method export and
// more advanced authentication mechanism, such as oauth 2.0 client credentials
type HTTPSender struct {
	config               HTTPSenderConfig
	secretsLastRetrieved time.Time
	cacheSecrets         map[string]string
	client               *http.Client
	mutex                sync.Mutex
}

func (sender *HTTPSender) createClient(ctx interfaces.AppFunctionContext) (*http.Client, error) {
	client := &http.Client{}

	// To deal with scenarios where both self-signed and CA certificates are used, always retrieve host's root CA set.
	// e.g. oauth2 token server uses CA-signed certificate and target https server uses Self-Signed certificate or vice versa,
	// get the system CA pool(x509.SystemCertPool()) instead of creating an empty CA pool(x509.NewCertPool())
	rootCASets, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("fail to get system root CA set: %s", err.Error())
	}

	tlsConfig := &tls.Config{
		RootCAs:            rootCASets,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: sender.config.SkipCertVerify, //nolint: gosec
		Renegotiation:      sender.config.RenegotiationSupport,
	}

	// For cases that target server uses self-signed cert, users must provide the self-signed cert into secret
	// provider with secret key "cacert". Following codes append such self-signed cert into the TLSClientConfig
	// of HTTP Client Transport.  Note that XpertHTTPExport only appends cacert when there is cacert secret specified
	// in secret provider, which is not relevant to the value of AuthMode.(authMode can be none and export to an HTTPS URL)
	if cacert, ok := sender.cacheSecrets[messaging.SecretCACert]; ok && len(cacert) > 0 {
		ctx.LoggingClient().Debugf("non-empty secret(%s) is specified, prepare to append this certificate into the CA cert pool", messaging.SecretCACert)
		// append the cacert as found in the secret provider
		if !tlsConfig.RootCAs.AppendCertsFromPEM([]byte(cacert)) {
			return nil, fmt.Errorf("fail to append %s as found in the secret provider into the CA cert pool", messaging.SecretCACert)
		}
	}
	client.Transport = &http.Transport{TLSClientConfig: tlsConfig}

	ctx.LoggingClient().Debugf("prepare HTTP client per authentication mode:%s", sender.config.AuthMode)
	switch sender.config.AuthMode {
	case HTTPAuthModeNONE:
		return client, nil
	case HTTPAuthModeHeaderSecret:
		return sender.prepareHTTPClientForHeaderSecretAuth(ctx, client)
	case HTTPAuthModeOauth2ClientCredentials:
		// if the oauth2 token server uses self-signed certificate, users shall provide such certificate via "token_cert" secret
		if tokenCert, ok := sender.cacheSecrets[SecretKeyTokenServerCert]; ok && len(tokenCert) > 0 {
			ctx.LoggingClient().Debugf("non-empty secret(%s) is specified, prepare to append this certificate into the CA cert pool", SecretKeyTokenServerCert)
			if !tlsConfig.RootCAs.AppendCertsFromPEM([]byte(tokenCert)) {
				return nil, fmt.Errorf("fail to append %s as found in the secret provider into the CA cert pool", SecretKeyTokenServerCert)
			}
		}
		return sender.prepareHTTPClientForOAuth2ClientCredentialsAuth(ctx, client)
	case HTTPAuthModeClientCert:
		return sender.prepareHTTPClientForClientCertAuth(ctx, client, tlsConfig)
	case HTTPAuthModeAWSSignature:
		return sender.prepareHTTPClientForAWSSignatureAuth(ctx, client)
	default:
		return nil, fmt.Errorf("unknown AuthMode %s for XpertHTTPExport", sender.config.AuthMode)
	}
}

func (sender *HTTPSender) initialize(ctx interfaces.AppFunctionContext) error {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()
	// If the conditions changed while waiting for the lock, i.e. other thread completed the initialization,
	// then skip doing anything
	if sender.client != nil && !sender.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		return nil
	}

	// only retrieve secrets from secretProvider when secretPath is specified
	if len(sender.config.SecretPath) > 0 {
		secrets, err := ctx.GetSecret(sender.config.SecretPath)
		if err != nil {
			return fmt.Errorf("fail to retrieve secrets inside XpertHTTPExport. Error: %s", err)
		}
		sender.cacheSecrets = secrets
	} else {
		sender.cacheSecrets = make(map[string]string)
	}
	sender.secretsLastRetrieved = time.Now()

	client, err := sender.createClient(ctx)
	if err != nil {
		return err
	}
	sender.client = client

	return nil
}

// HTTPSend will send data from the previous function to the specified Endpoint via http requests with supported
// method GET, POST, PUT, PATCH, or DELETE and various authentication mechanism. If no previous function exists, the
// event that triggered the pipeline will be used. Passing an empty string to the mimetype method will default to
// application/json.
func (sender *HTTPSender) HTTPSend(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()
	lc.Debugf("HTTP Exporting in pipeline '%s'", ctx.PipelineId())

	if data == nil {
		// No data received
		return false, fmt.Errorf("XpertHTTPExport function in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	exportData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}

	// check if client haven't been initialized OR the cache has been invalidated (due to new/updated secrets), need to (re)initialize the client
	if sender.client == nil || sender.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		err = sender.initialize(ctx)
		if err != nil {
			return false, err
		}
	}

	req, err := sender.prepareHTTPRequest(ctx, exportData)
	if err != nil {
		return false, err
	}

	ctx.LoggingClient().Debugf("sending HTTP %s request to %s in pipeline '%s'", req.Method, req.URL.String(), ctx.PipelineId())
	response, err := sender.client.Do(req)
	// Pipeline continues if we get a 2xx response, non-2xx response may stop pipeline
	if err != nil || response.StatusCode < 200 || response.StatusCode >= 300 {
		if err == nil {
			err = fmt.Errorf("XpertHTTPExport received %d HTTP response status code after sending %s request to %s in pipeline '%s'", response.StatusCode, req.Method, req.URL.String(), ctx.PipelineId())
			body, readErr := io.ReadAll(response.Body)
			if readErr == nil {
				ctx.LoggingClient().Errorf("HTTP request non-2xx response body : %s", string(body))
			}
			defer response.Body.Close()
		} else {
			err = fmt.Errorf("XpertHTTPExport received error after sending %s request to %s in pipeline '%s': %s", req.Method, req.URL.String(), ctx.PipelineId(), err.Error())
		}

		// If continuing on send error then can't be persisting on error since Store and Forward retries starting
		// with the function that failed and stopped the execution of the pipeline.  Note that if both
		// continueOnSendError and persistOnError are true, logic here makes continueOnSendError=true take precedence
		// over persistOnError=true.
		if !sender.config.ContinueOnSendError {
			sender.setRetryData(ctx, exportData)
			return false, err
		}

		// Continuing pipeline on error
		// This is in support of sending to multiple export destinations by chaining export functions in the pipeline.
		ctx.LoggingClient().Errorf("Continuing pipeline on error in pipeline '%s': %s", ctx.PipelineId(), err.Error())

		// Return the input data since must have some data for the next function to operate on.
		return true, data
	}

	// capture the size into metrics
	exportDataBytes := len(exportData)
	// TODO: EdgeX 3.0 refactor size metrics once receivers are pointers (like mqtt size metrics)
	metrics := gometrics.DefaultRegistry.Get(internal.HttpExportSizeName)
	var httpExportSizeMetric gometrics.Histogram
	if metrics == nil {
		var err error
		lc.Debugf("Initializing metric %s.", internal.HttpExportSizeName)
		httpExportSizeMetric = gometrics.NewHistogram(gometrics.NewUniformSample(internal.MetricsReservoirSize))
		metricsManger := ctx.MetricsManager()
		if metricsManger != nil {
			// TODO: EdgeX 3.0 append url to export size name
			err = metricsManger.Register(internal.HttpExportSizeName, httpExportSizeMetric, nil)
		} else {
			err = errors.New("metrics manager not available")
		}

		if err != nil {
			lc.Errorf("Unable to register metric %s. Collection will continue, but metric will not be reported: %s", internal.HttpExportSizeName, err.Error())
		}

	} else {
		httpExportSizeMetric = metrics.(gometrics.Histogram)
	}
	httpExportSizeMetric.Update(int64(exportDataBytes))

	ctx.LoggingClient().Debugf("Sent %d bytes of data in pipeline '%s'. Response status is %s", len(exportData), ctx.PipelineId(), response.Status)
	ctx.LoggingClient().Tracef("Data exported for pipeline '%s' (%s=%s)", ctx.PipelineId(), common.CorrelationHeader, ctx.CorrelationID())

	// This allows multiple HTTP Exports to be chained in the pipeline to send the same data to different destinations
	// Don't need to read the response data since not going to return it so just return now.
	if sender.config.ReturnInputData {
		return true, data
	}

	defer func() { _ = response.Body.Close() }()
	responseData, errReadingBody := io.ReadAll(response.Body)
	if errReadingBody != nil {
		// Can't have continueOnSendError=true when returnInputData=false, so no need to check for it here
		sender.setRetryData(ctx, exportData)
		return false, errReadingBody
	}

	return true, responseData
}

func (sender *HTTPSender) prepareHTTPRequest(ctx interfaces.AppFunctionContext, data []byte) (*http.Request, error) {
	formattedUrl, err := ctx.ApplyValues(sender.config.URL)
	if err != nil {
		return nil, err
	}

	parsedUrl, err := url.Parse(formattedUrl)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(sender.config.HTTPMethod, parsedUrl.String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set(common.ContentType, sender.config.MimeType)
	// per http://craigwickesser.com/2015/01/golang-http-to-many-open-files/, setting the Close property to true to
	// avoid "too many open files" error
	req.Close = true

	if sender.config.AuthMode == HTTPAuthModeHeaderSecret {
		req.Header.Set(sender.config.HTTPHeaderName, sender.cacheSecrets[sender.config.SecretName])
	}

	if sender.config.AuthMode == HTTPAuthModeAWSSignature {
		err := sender.prepareAWSSignedRequest(ctx, req, data)
		if err != nil {
			return nil, err
		}
	}

	if len(sender.config.HTTPRequestHeaders) != 0 {
		for k, v := range sender.config.HTTPRequestHeaders {
			req.Header.Set(k, v)
		}
	}

	return req, nil
}

func (sender *HTTPSender) prepareHTTPClientForHeaderSecretAuth(ctx interfaces.AppFunctionContext, client *http.Client) (*http.Client, error) {
	ctx.LoggingClient().Debug("validate if mandatory secrets exist")
	if len(sender.config.HTTPHeaderName) == 0 {
		return nil, fmt.Errorf("HTTP Header Name is required when using AuthMode %s", HTTPAuthModeHeaderSecret)
	}
	if len(sender.config.SecretPath) == 0 {
		return nil, fmt.Errorf("secretPath is required when using AuthMode %s", HTTPAuthModeHeaderSecret)
	}
	if len(sender.config.SecretName) == 0 {
		return nil, fmt.Errorf("secretName is required when using AuthMode %s", HTTPAuthModeHeaderSecret)
	}
	if _, ok := sender.cacheSecrets[sender.config.SecretName]; !ok {
		return nil, fmt.Errorf("no corresponding secret can be found with secretName %s when using AuthMode %s", sender.config.SecretName, HTTPAuthModeHeaderSecret)
	}
	return client, nil
}

func (sender *HTTPSender) prepareHTTPClientForOAuth2ClientCredentialsAuth(ctx interfaces.AppFunctionContext, client *http.Client) (*http.Client, error) {
	ctx.LoggingClient().Debugf("validate if mandatory secrets exist")
	err := sender.validateSecrets(sender.cacheSecrets, SecretKeyClientId, SecretKeyClientSecret, SecretKeyTokenUrl)
	if err != nil {
		return nil, err
	}
	urlValues := url.Values{}
	conf := &clientcredentials.Config{}
	for k, v := range sender.cacheSecrets {
		switch k {
		case SecretKeyClientId:
			conf.ClientID = v
		case SecretKeyClientSecret:
			conf.ClientSecret = v
		case SecretKeyTokenScopes:
			scopes := strings.Split(v, ",")
			conf.Scopes = scopes
		case SecretKeyTokenUrl:
			conf.TokenURL = v
		case SecretKeyTokenServerCert, messaging.SecretCACert:
			// skip certificate secrets as they're for self-signed CA certs and are not relevant to oauth2
		default:
			urlValues.Set(k, v)
		}
	}
	conf.EndpointParams = urlValues
	oauth2Client := conf.Client(context.WithValue(context.TODO(), oauth2.HTTPClient, client))
	ctx.LoggingClient().Debugf("created HTTP client that will obtain OAuth2 token from %s", conf.TokenURL)
	return oauth2Client, nil
}

func (sender *HTTPSender) prepareHTTPClientForClientCertAuth(ctx interfaces.AppFunctionContext, client *http.Client,
	tlsConfig *tls.Config) (*http.Client, error) {
	ctx.LoggingClient().Debug("validate if mandatory secrets exist")
	if len(sender.config.SecretPath) == 0 {
		return nil, fmt.Errorf("secretPath is required when using AuthMode %s", HTTPAuthModeClientCert)
	}
	var clientKey, clientCert string
	var ok bool
	if clientKey, ok = sender.cacheSecrets[messaging.SecretClientKey]; !ok {
		return nil, fmt.Errorf("no corresponding secret can be found with secretName %s when using AuthMode %s",
			messaging.SecretClientKey, HTTPAuthModeClientCert)
	}
	if clientCert, ok = sender.cacheSecrets[messaging.SecretClientCert]; !ok {
		return nil, fmt.Errorf("no corresponding secret can be found with secretName %s when using AuthMode %s",
			messaging.SecretClientCert, HTTPAuthModeClientCert)
	}
	cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
	if err != nil {
		return nil, fmt.Errorf("error parsing client cert: %v", err)
	}
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	client.Transport = &http.Transport{TLSClientConfig: tlsConfig}
	return client, nil
}

func (sender *HTTPSender) prepareHTTPClientForAWSSignatureAuth(ctx interfaces.AppFunctionContext, client *http.Client) (*http.Client, error) {
	ctx.LoggingClient().Debugf("validate if mandatory secrets %s, %s exist", SecretKeyAWSAccessKey, SecretKeyAWSSecretKey)
	err := sender.validateSecrets(sender.cacheSecrets, SecretKeyAWSAccessKey, SecretKeyAWSSecretKey)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (sender *HTTPSender) prepareAWSSignedRequest(appCtx interfaces.AppFunctionContext, req *http.Request, data []byte) error {
	// context in awsSigner is only for the optional context aware log entries
	// to make it useful, we need to implement the Logger and ContextLogger interfaces to wrap loggingClient
	// https://github.com/aws/smithy-go/blob/7f8b3b99628aa810894bd866caa1998421c807e0/logging/logger.go#L19
	// and pass it as a SignerOptions with LogSigning = true
	ctx := context.TODO()
	hash := sha256.New()
	hash.Write(data)
	hashData := hex.EncodeToString(hash.Sum(nil))
	credentials := aws.Credentials{AccessKeyID: sender.cacheSecrets[SecretKeyAWSAccessKey], SecretAccessKey: sender.cacheSecrets[SecretKeyAWSSecretKey]}

	awsSigner := awsv4signer.NewSigner()
	if strings.EqualFold(sender.config.AWSV4SignerConfigs[AWSV4SignerConfigsService], "s3") {
		// Set S3 required header "X-Amz-Content-Sha256" because it is not set by SignHTTP
		req.Header.Set("X-Amz-Content-Sha256", hashData)
	}
	err := awsSigner.SignHTTP(ctx, credentials, req, hashData, sender.config.AWSV4SignerConfigs[AWSV4SignerConfigsService], sender.config.AWSV4SignerConfigs[AWSV4SignerConfigsRegion], time.Now())
	appCtx.LoggingClient().Debugf("AWS signed request %v with headers %v", (*req).URL, (*req).Header)

	return err
}

func (sender *HTTPSender) validateSecrets(secrets map[string]string, keys ...string) error {
	for _, key := range keys {
		if value, ok := secrets[key]; !ok || len(value) == 0 {
			return fmt.Errorf("secret %s doesn't exist or is empty", key)
		}
	}
	return nil
}

func (sender *HTTPSender) setRetryData(ctx interfaces.AppFunctionContext, exportData []byte) {
	if sender.config.PersistOnError {
		ctx.SetRetryData(exportData)
	}
}
