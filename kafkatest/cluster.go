// Inspired and adapted from a testcontainers-go PR:
// https://github.com/testcontainers/testcontainers-go/pull/356
// In turn inspired by Java Kafka testcontainers' module
// https://github.com/testcontainers/testcontainers-java/blob/master/modules/kafka/src/main/java/org/testcontainers/containers/KafkaContainer.java

package kafkatest

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"software.sslmate.com/src/go-pkcs12"
)

const (
	zookeeperPort   = 2181
	kafkaBrokerPort = 9092
)

type KafkaCluster struct {
	network     testcontainers.Network
	networkName string

	zookeeperContainer testcontainers.Container
	kafkaContainer     testcontainers.Container

	clientTLS *tls.Config
	tlsPhrase string
	options   Options
}

// NewKafkaClusterWithOptions creates and starts a new kafka cluster via the given Options object.
func NewKafkaClusterWithOptions(ctx context.Context, origOptions *Options) (*KafkaCluster, error) {
	options := *origOptions
	if options.Logger == nil {
		options.Logger = log.New(io.Discard, "", log.LstdFlags)
	}

	networkName := options.NetworkPrefix + "-" + uuid.NewString()
	options.Logger.Printf("Using %q as kafka cluster network", networkName)

	// creates a network, so kafka and zookeeper can communicate directly
	networkCtx, cancel := ctxWithTimeout(ctx, options.NetworkCreateTimeout)
	defer cancel()
	network, err := testcontainers.GenericNetwork(networkCtx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{Name: networkName, CheckDuplicate: true},
	})
	if err != nil {
		return nil, err
	}

	options.Logger.Printf("Kafka cluster network created")

	kc := &KafkaCluster{
		tlsPhrase: uuid.NewString(),
		clientTLS: nil,

		network:     network,
		networkName: networkName,

		options: options,
	}

	if err := kc.createZookeeperContainer(ctx); err != nil {
		return nil, err
	}

	if err := kc.createKafkaContainer(ctx); err != nil {
		return nil, err
	}

	if err := kc.startCluster(ctx); err != nil {
		return nil, err
	}

	return kc, nil
}

// NewKafkaCluster creates and starts a new kafka cluster via the given Option objects.
func NewKafkaCluster(ctx context.Context, setters ...Option) (*KafkaCluster, error) {

	opts := &Options{
		ZookeeperImage:  DefaultZookeeperImage,
		KafkaImage:      DefaultKafkaImage,
		KafkaClientPort: DefaultKafkaClientPort,

		NetworkPrefix: DefaultNetworkPrefix,
		TLS:           DefaultTLS,

		ZookeeperWaitStrategy: DefaultZookeeperWaitStrategy,
		KafkaWaitStrategy:     DefaultKafkaWaitStrategy,

		ContainerStartTimeout:    DefaultContainerStartTimeout,
		ContainerShutdownTimeout: DefaultContainerShutdownTimeout,
		NetworkCreateTimeout:     DefaultNetworkCreateTimeout,
		NetworkShutdownTimeout:   DefaultNetworkShutdownTimeout,

		Logger: DefaultLogger,
	}

	for _, setter := range setters {
		setter(opts)
	}

	return NewKafkaClusterWithOptions(ctx, opts)
}

func (kc *KafkaCluster) startCluster(ctx context.Context) error {
	zkCtx, cancel := ctxWithTimeout(ctx, kc.options.ContainerStartTimeout)
	defer cancel()
	if err := kc.zookeeperContainer.Start(zkCtx); err != nil {
		return err
	}

	kafkaCtx, cancel := ctxWithTimeout(ctx, kc.options.ContainerStartTimeout)
	defer cancel()
	if err := kc.kafkaContainer.Start(kafkaCtx); err != nil {
		return err
	}

	return kc.configureKafka(kafkaCtx)
}

// StopCluster gracefully stops the cluster.
func (kc *KafkaCluster) StopCluster(ctx context.Context) {
	zkCtx, cancel := ctxWithTimeout(ctx, kc.options.ContainerShutdownTimeout)
	defer cancel()

	if err := kc.zookeeperContainer.Stop(zkCtx, nil); err != nil {
		kc.options.Logger.Printf("failed to shutdown zookeeper container: %s", err)
	}

	kafkaCtx, cancel := ctxWithTimeout(ctx, kc.options.ContainerShutdownTimeout)
	defer cancel()

	if err := kc.kafkaContainer.Stop(kafkaCtx, nil); err != nil {
		kc.options.Logger.Printf("failed to shutdown kafka container: %s", err)
	}

	ntCtx, cancel := ctxWithTimeout(ctx, kc.options.NetworkShutdownTimeout)
	defer cancel()

	if err := kc.network.Remove(ntCtx); err != nil {
		kc.options.Logger.Printf("failed to remove cluster network: %s", err)
	}
}

func (kc *KafkaCluster) configureKafka(ctx context.Context) error {
	kc.options.Logger.Printf("Configuring kafka container")

	// Configure WithTLS things
	if kc.options.TLS {
		kafkaHost, err := kc.GetKafkaHost(ctx)
		if err != nil {
			return err
		}

		ca, brokerCert, clientTLS, err := generateKafkaTLS(kafkaHost)
		if err != nil {
			return err
		}

		if err := kc.copyTLS(ctx, ca, brokerCert); err != nil {
			return err
		}

		kc.clientTLS = clientTLS
	}

	// Set KAFKA_ADVERTISED_LISTENERS
	exposedHost, err := kc.GetKafkaHostAndPort(ctx)
	if err != nil {
		return err
	}

	// needs to set KAFKA_ADVERTISED_LISTENERS with the exposed kafka port
	kafkaStartFile := strings.Builder{}
	kafkaStartFile.WriteString("#!/bin/bash \n")
	kafkaStartFile.WriteString(fmt.Sprintf("export KAFKA_ADVERTISED_LISTENERS='EXTERNAL://%s,BROKER://kafka:%d'\n", exposedHost, kafkaBrokerPort))
	kafkaStartFile.WriteString("/etc/confluent/docker/run \n")

	if err := kc.kafkaContainer.CopyToContainer(ctx, []byte(kafkaStartFile.String()), "/testcontainers_start.sh", 0777); err != nil {
		return err
	}

	kc.options.Logger.Printf("Waiting for kafka container to be completely up")

	// The wait needs to happen here, since only when /testcontainers_start.sh is copied, the container
	// can actually start.
	if err := kc.options.KafkaWaitStrategy.WaitUntilReady(ctx, kc.kafkaContainer); err != nil {
		return err
	}

	kc.options.Logger.Printf("Kafka container ready")

	return nil
}

// GetKafkaHostAndPort gets the kafka host:port, so it can be accessed from outside the container.
func (kc *KafkaCluster) GetKafkaHostAndPort(ctx context.Context) (string, error) {
	host, err := kc.GetKafkaHost(ctx)
	if err != nil {
		return "", err
	}

	port, err := kc.kafkaContainer.MappedPort(
		ctx,
		nat.Port(fmt.Sprintf("%d/tcp", kc.options.KafkaClientPort)),
	)
	if err != nil {
		return "", err
	}

	// returns the exposed kafka host:port
	return host + ":" + port.Port(), nil
}

// GetKafkaHost gets the kafka host, so it can be accessed from outside the container.
func (kc *KafkaCluster) GetKafkaHost(ctx context.Context) (string, error) {
	return kc.kafkaContainer.Host(ctx)
}

func (kc *KafkaCluster) createZookeeperContainer(ctx context.Context) error {
	// creates the zookeeper container, but do not start it yet
	zookeeperContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        kc.options.ZookeeperImage,
			ExposedPorts: []string{fmt.Sprintf("%d/tcp", zookeeperPort)},
			Env: map[string]string{
				"ZOOKEEPER_CLIENT_PORT": strconv.Itoa(zookeeperPort),
				"ZOOKEEPER_TICK_TIME":   "2000",
			},
			Networks:       []string{kc.networkName},
			NetworkAliases: map[string][]string{kc.networkName: {"zookeeper"}},
			WaitingFor:     kc.options.ZookeeperWaitStrategy,
		},
	})
	if err != nil {
		return err
	}

	kc.zookeeperContainer = zookeeperContainer
	kc.options.Logger.Printf("Zookeeper container with ID %q created", zookeeperContainer.GetContainerID())

	return nil
}

func (kc *KafkaCluster) createKafkaContainer(ctx context.Context) error {
	// creates the kafka container, but do not start it yet

	parameters := map[string]string{
		"KAFKA_BROKER_ID":                        "1",
		"KAFKA_ZOOKEEPER_CONNECT":                fmt.Sprintf("zookeeper:%d", zookeeperPort),
		"KAFKA_LISTENERS":                        fmt.Sprintf("EXTERNAL://0.0.0.0:%d,BROKER://0.0.0.0:%d", kc.options.KafkaClientPort, kafkaBrokerPort),
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "BROKER:PLAINTEXT,EXTERNAL:PLAINTEXT",
		"KAFKA_INTER_BROKER_LISTENER_NAME":       "BROKER",
		"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
	}

	if kc.options.TLS {
		tlsParameters := map[string]string{
			// Broker SSL encryption and authentication
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "BROKER:PLAINTEXT,EXTERNAL:SSL",
			"KAFKA_SSL_KEYSTORE_LOCATION":          "/kafka.broker.keypair.pkcs12",
			"KAFKA_SSL_KEYSTORE_TYPE":              "PKCS12",
			"KAFKA_SSL_TRUSTSTORE_LOCATION":        "/ca.crt",
			"KAFKA_SSL_TRUSTSTORE_TYPE":            "PEM",
			"KAFKA_SSL_KEY_PASSWORD":               kc.tlsPhrase,
			"KAFKA_SSL_KEYSTORE_PASSWORD":          kc.tlsPhrase,
		}

		for k, v := range tlsParameters {
			parameters[k] = v
		}
	}

	//wat := nat.Port(fmt.Sprintf("%d/tcp", kc.options.WithKafkaClientPort))
	kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:          kc.options.KafkaImage,
			ExposedPorts:   []string{fmt.Sprintf("%d/tcp", kc.options.KafkaClientPort)},
			Env:            parameters,
			Networks:       []string{kc.networkName},
			NetworkAliases: map[string][]string{kc.networkName: {"kafka"}},
			Cmd:            []string{"/bin/sh", "-c", "while [ ! -f /testcontainers_start.sh ]; do sleep 0.1; done; /testcontainers_start.sh"},
		},
	})
	if err != nil {
		return err
	}

	kc.kafkaContainer = kafkaContainer
	kc.options.Logger.Printf("Kafka container with ID %q created", kafkaContainer.GetContainerID())

	return nil
}

func (kc *KafkaCluster) copyTLS(ctx context.Context, ca *x509.Certificate, brokerCert tls.Certificate) error {
	// CA
	if err := kc.kafkaContainer.CopyToContainer(ctx, encodeCertificates(ca), "/ca.crt", 0444); err != nil {
		return err
	}

	// Keypair
	brokerCertificates, err := x509.ParseCertificate(brokerCert.Certificate[0])
	if err != nil {
		return err
	}

	pfxData, err := pkcs12.Encode(rand.Reader, brokerCert.PrivateKey, brokerCertificates, []*x509.Certificate{ca}, kc.tlsPhrase)
	if err != nil {
		return err
	}

	if err := kc.kafkaContainer.CopyToContainer(ctx, pfxData, "/kafka.broker.keypair.pkcs12", 0444); err != nil {
		return err
	}

	return nil
}

// ClientTLSConfig returns the tls.Config which can be used by clients to enable WithTLS encryption with the cluster,
// if configured.
func (kc *KafkaCluster) ClientTLSConfig() *tls.Config {
	return kc.clientTLS
}

// ctxWithTimeout wraps the given parent context.Context, if the timeout is > 0. Otherwise,
// the parent context.Context is returned as-is, with a nop context.CancelFunc.
func ctxWithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return parent, func() {}
	}

	return context.WithTimeout(parent, timeout)
}

func encodeCertificates(certificates ...*x509.Certificate) []byte {
	var byteBuffer []byte
	for _, certificate := range certificates {
		byteBuffer = append(byteBuffer, pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certificate.Raw,
		})...)
	}

	return byteBuffer
}

func generateKafkaTLS(kafkaExternalName string) (*x509.Certificate, tls.Certificate, *tls.Config, error) {
	// Inspired by:
	// https://shaneutt.com/blog/golang-ca-and-signed-cert-go/

	caCert, caKey, err := generateCA()
	if err != nil {
		return nil, tls.Certificate{}, nil, err
	}

	brokerCert, err := generateCertificate(kafkaExternalName, caCert, caKey)
	if err != nil {
		return nil, tls.Certificate{}, nil, err
	}

	caPool := x509.NewCertPool()
	caPool.AddCert(caCert)

	return caCert, brokerCert, &tls.Config{
		RootCAs: caPool,
	}, nil
}

func generateCA() (*x509.Certificate, crypto.PrivateKey, error) {
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(2019),
		Subject:               pkix.Name{CommonName: "Kafka-Security-CA"},
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	caCertBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	caCert, err := x509.ParseCertificate(caCertBytes)
	if err != nil {
		return nil, nil, err
	}

	return caCert, caPrivKey, nil
}

func generateCertificate(dnsTarget string, ca *x509.Certificate, caPrivKey crypto.PrivateKey) (tls.Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{dnsTarget, "kafka"},
		NotAfter:     time.Now().AddDate(10, 0, 0),
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return tls.Certificate{}, nil
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return tls.Certificate{}, nil
	}

	certPrivKeyBytes, err := x509.MarshalPKCS8PrivateKey(certPrivKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.X509KeyPair(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}), pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: certPrivKeyBytes,
	}))
}
