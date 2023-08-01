// Copyright (C) 2021-2022 IOTech Ltd

package xpert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
)

const (
	KafkaSecretKeyPem    = messaging.SecretClientKey
	KafkaSecretCertPem   = messaging.SecretClientCert
	KafkaSecretCaCertPem = messaging.SecretCACert
	KafkaDecryptPassword = "decryptedpassword"
	RSA_PRIVATE_KEY_TYPE = "RSA PRIVATE KEY"
)

type KafkaEndpoint struct {
	// ClientID for logging, debugging and auditing purposes
	ClientID string
	// The Kafka broker address
	Address string
	// The port number of the Kafka broker
	Port int
	// The Kafka topic to which the message is sent
	Topic string
	// The partition to which the message is sent
	Partition int32
}

type KafkaSecretsConfig struct {
	// The name of the path in secret provider to retrieve your secrets
	SecretPath string
	// AuthMode specifies what authentication mode to produce messages to Kafka broker
	AuthMode string
	// SkipCertVerify specifies whether the Edge Xpert Application Service verifies the server's certificate chain and host name
	SkipCertVerify bool
}

type kafkaSender struct {
	producer             sarama.SyncProducer
	config               *sarama.Config
	endpoint             KafkaEndpoint
	secretsConfig        KafkaSecretsConfig
	secretsLastRetrieved time.Time
	mutex                sync.Mutex
	persistOnError       bool
}

type kafkaSecrets struct {
	keyPEMBlock       []byte
	certPEMBlock      []byte
	caCertPEMBlock    []byte
	decryptedPassword []byte
}

// NewKafkaSender - create new kafka sender
func NewKafkaSender(kafkaEndpoint KafkaEndpoint, secretConfig KafkaSecretsConfig, persistOnError bool) *kafkaSender {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.ClientID = kafkaEndpoint.ClientID

	sender := &kafkaSender{
		config:         config,
		endpoint:       kafkaEndpoint,
		secretsConfig:  secretConfig,
		persistOnError: persistOnError,
	}

	return sender
}

func (sender *kafkaSender) getSecrets(ctx interfaces.AppFunctionContext) (*kafkaSecrets, error) {
	secrets, err := ctx.GetSecret(sender.secretsConfig.SecretPath)
	if err != nil {
		return nil, err
	}
	keyPEMBlock := []byte(secrets[KafkaSecretKeyPem])
	certPEMBlock := []byte(secrets[KafkaSecretCertPem])
	caCertPEMBlock := []byte(secrets[KafkaSecretCaCertPem])
	password := []byte(secrets[KafkaDecryptPassword])
	kafkaSecrets := &kafkaSecrets{
		keyPEMBlock:       keyPEMBlock,
		certPEMBlock:      certPEMBlock,
		caCertPEMBlock:    caCertPEMBlock,
		decryptedPassword: password,
	}
	// attempt to decrypt client key in case that user may specify encrypted key
	if err := sender.decryptClientKey(kafkaSecrets); err != nil {
		return nil, err
	}
	return kafkaSecrets, nil
}

func (sender *kafkaSender) validateSecrets(secrets kafkaSecrets) error {
	switch sender.secretsConfig.AuthMode {
	case messaging.AuthModeCert:
		// need both to make a successful connection
		if len(secrets.keyPEMBlock) <= 0 || len(secrets.certPEMBlock) <= 0 {
			return fmt.Errorf("%s authentication mode selected however the key or cert PEM block was not found at secret path", messaging.AuthModeCert)
		}
		// cacert is optional, and only verified when users specify the cacert
		if len(secrets.caCertPEMBlock) > 0 {
			caCertPool := x509.NewCertPool()
			ok := caCertPool.AppendCertsFromPEM(secrets.caCertPEMBlock)
			if !ok {
				return errors.New("error parsing CA Certificate")
			}
		}
	}
	return nil
}

func (sender *kafkaSender) decryptClientKey(secrets *kafkaSecrets) error {
	// only try to decrypt the private key when decryptedPassword is specified
	if len(secrets.decryptedPassword) > 0 {
		//decode the byte array to find out the first pem block
		decodedPem, _ := pem.Decode(secrets.keyPEMBlock)
		if decodedPem == nil {
			return errors.New("client key is not in PEM format")
		}
		// only deal with RSA private key and determine if it's encrypted PEM block
		// Both x509.IsEncryptedPEMBlock and x509.DecryptPEMBlock have been deprecated since go 1.16.
		// However, go standard libraries doesn't provide alternative function to decrypt PEM block with password.
		// As a result, we will keep the usage of both deprecated functions until alternative functions are available.
		// To avoid lint warning, mark the usage of both deprecated functions with //nolint: staticcheck
		if decodedPem.Type == RSA_PRIVATE_KEY_TYPE && x509.IsEncryptedPEMBlock(decodedPem) { //nolint: staticcheck
			// decrypt the encrypted PEM block by using password specified by users
			pkey, err := x509.DecryptPEMBlock(decodedPem, secrets.decryptedPassword) //nolint: staticcheck
			if err != nil {
				return fmt.Errorf("fail to decrypt client key. error: %v", err)
			}
			// reset the private key pem block to the decrypted one
			secrets.keyPEMBlock = pem.EncodeToMemory(
				&pem.Block{
					Type:  decodedPem.Type,
					Bytes: pkey,
				})
		}
	}
	return nil
}

func (sender *kafkaSender) configureKafkaProducerForAuth(secrets kafkaSecrets) error {
	// validate secrets prior to configure TLS
	err := sender.validateSecrets(secrets)
	if err != nil {
		return err
	}
	var cert tls.Certificate
	cert, err = tls.X509KeyPair(secrets.certPEMBlock, secrets.keyPEMBlock)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: sender.secretsConfig.SkipCertVerify, //nolint: gosec
		Certificates:       []tls.Certificate{cert},
	}

	// if cacert is specified, set cacert to RootCAs
	if len(secrets.caCertPEMBlock) > 0 {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(secrets.caCertPEMBlock)
		tlsConfig.RootCAs = caCertPool
	}
	sender.config.Net.TLS.Enable = true
	sender.config.Net.TLS.Config = tlsConfig

	return nil
}

func (sender *kafkaSender) initializeKafkaSyncProducer(ctx interfaces.AppFunctionContext) error {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	// If the conditions changed while waiting for the lock, i.e. other thread completed the initialization,
	// then skip doing anything
	if sender.producer != nil && !sender.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		return nil
	}
	if sender.secretsConfig.AuthMode != messaging.AuthModeNone {
		//get the secrets from the secret provider and populate the struct
		secrets, err := sender.getSecrets(ctx)
		if err != nil {
			return err
		}
		//ensure that the authmode selected has the required secret values
		if secrets != nil {
			// validate secrets prior to configure TLS
			err := sender.configureKafkaProducerForAuth(*secrets)
			if err != nil {
				return err
			}
		}
	}
	kafkaServerURL := fmt.Sprintf("%s:%d", sender.endpoint.Address, sender.endpoint.Port)
	ctx.LoggingClient().Infof("creating Kafka producer for endpoint: %s", kafkaServerURL)
	// Create the producer
	kafkaProducer, err := sarama.NewSyncProducer([]string{kafkaServerURL}, sender.config)
	if err != nil {
		return fmt.Errorf("could not create Kafka producer, %s. Error: %s", kafkaServerURL, err)
	}
	ctx.LoggingClient().Infof("Kafka producer has been successfully created for endpoint: %s", kafkaServerURL)
	sender.producer = kafkaProducer
	// Once producer is successfully initialized, secretsLastRetrieved needs to be updated no matter what the authMode is,
	// as secretsLastRetrieved comparison is part of condition to decide if producer should be initialized.
	sender.secretsLastRetrieved = time.Now()
	return nil
}

func (sender *kafkaSender) KafkaSend(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		// We didn't receive a result
		return false, errors.New("no data received")
	}

	exportData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}

	// check if client haven't be initialized OR the cache has been invalidated (due to new/updated secrets), need to (re)initialize the client
	if sender.producer == nil || sender.secretsLastRetrieved.Before(ctx.SecretsLastUpdated()) {
		ctx.LoggingClient().Info("Initializing Kafka producer...")
		err := sender.initializeKafkaSyncProducer(ctx)
		if err != nil {
			return false, err
		}
	}

	msg := &sarama.ProducerMessage{
		Topic:     sender.endpoint.Topic,
		Partition: sender.endpoint.Partition,
		Value:     sarama.ByteEncoder(exportData),
	}
	partition, offset, err := sender.producer.SendMessage(msg)
	if err != nil {
		sender.setRetryData(ctx, exportData)
		subMessage := "drop event"
		if sender.persistOnError {
			subMessage = "persisting Event for later retry"
		}
		return false, fmt.Errorf("failed to send message from kafka producer, %s. Error: %s", subMessage, err)
	}
	ctx.LoggingClient().Infof("Message sent to partition %d @ Offset %d", partition, offset)
	return true, nil
}

func (sender *kafkaSender) setRetryData(ctx interfaces.AppFunctionContext, exportData []byte) {
	if sender.persistOnError {
		ctx.SetRetryData(exportData)
	}
}
