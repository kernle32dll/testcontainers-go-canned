package kafkatest

import (
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// DefaultZookeeperImage is the default image for the zookeeper container.
	DefaultZookeeperImage = "confluentinc/cp-zookeeper:7.2.2"

	// DefaultKafkaImage is the default image for the kafka container.
	DefaultKafkaImage = "confluentinc/cp-kafka:7.2.2"

	// DefaultKafkaClientPort is the default exposed port of the kafka container.
	DefaultKafkaClientPort = 9093

	// DefaultNetworkPrefix is the default prefix for the network used to connect containers.
	DefaultNetworkPrefix = "kafka-cluster"

	// DefaultTLS is the default for weather or not the cluster should use encrypted connections.
	DefaultTLS = false

	// DefaultContainerStartTimeout is the default timeout for starting containers
	DefaultContainerStartTimeout = time.Minute

	// DefaultContainerShutdownTimeout is the default timeout for shutting down containers
	DefaultContainerShutdownTimeout = time.Minute

	// DefaultNetworkCreateTimeout is the default timeout for creating the container network
	DefaultNetworkCreateTimeout = time.Minute

	// DefaultNetworkShutdownTimeout is the default timeout for shutting down the container network
	DefaultNetworkShutdownTimeout = time.Minute
)

var (
	// DefaultZookeeperWaitStrategy is the default wait.Strategy for the zookeeper container.
	DefaultZookeeperWaitStrategy = wait.ForLog("binding to port")

	// DefaultKafkaWaitStrategy is the default wait.Strategy for the kafka container.
	DefaultKafkaWaitStrategy = wait.ForLog("Initialized broker")

	// DefaultLogger is the default logger.
	DefaultLogger = testcontainers.Logger
)

// Options bundles all available configuration
// properties for a KafkaCluster.
type Options struct {
	ZookeeperImage  string
	KafkaImage      string
	KafkaClientPort int

	NetworkPrefix string
	TLS           bool

	ZookeeperWaitStrategy wait.Strategy
	KafkaWaitStrategy     wait.Strategy

	ContainerStartTimeout    time.Duration
	ContainerShutdownTimeout time.Duration
	NetworkCreateTimeout     time.Duration
	NetworkShutdownTimeout   time.Duration

	Logger testcontainers.Logging
}

// Option represents an option for a Kafka cluster.
type Option func(*Options)

// WithZookeeperImage sets image for the zookeeper container.
// The default is DefaultZookeeperImage.
func WithZookeeperImage(image string) Option {
	return func(c *Options) {
		c.ZookeeperImage = image
	}
}

// WithKafkaImage sets image for the kafka container.
// The default is DefaultKafkaImage.
func WithKafkaImage(image string) Option {
	return func(c *Options) {
		c.KafkaImage = image
	}
}

// WithKafkaClientPort sets exposed port of the kafka container.
// The default is DefaultKafkaClientPort.
func WithKafkaClientPort(port int) Option {
	return func(c *Options) {
		c.KafkaClientPort = port
	}
}

// WithNetworkPrefix sets prefix for the network used to connect containers.
// The default is DefaultNetworkPrefix.
func WithNetworkPrefix(prefix string) Option {
	return func(c *Options) {
		c.NetworkPrefix = prefix
	}
}

// WithTLS sets weather or not the cluster should use encrypted connections.
// The default is DefaultTLS.
func WithTLS(tls bool) Option {
	return func(c *Options) {
		c.TLS = tls
	}
}

// WithZookeeperWaitStrategy sets the wait.Strategy for the zookeeper container.
// The default is DefaultZookeeperWaitStrategy.
func WithZookeeperWaitStrategy(strategy wait.Strategy) Option {
	return func(c *Options) {
		c.ZookeeperWaitStrategy = strategy
	}
}

// WithKafkaWaitStrategy sets the wait.Strategy for the kafka container.
// The default is DefaultKafkaWaitStrategy.
func WithKafkaWaitStrategy(strategy wait.Strategy) Option {
	return func(c *Options) {
		c.KafkaWaitStrategy = strategy
	}
}

// WithContainerStartTimeout sets the timeout for starting containers.
// The default is DefaultContainerStartTimeout.
func WithContainerStartTimeout(timeout time.Duration) Option {
	return func(c *Options) {
		c.ContainerStartTimeout = timeout
	}
}

// WithContainerShutdownTimeout sets the timeout for shutting down containers.
// The default is DefaultContainerShutdownTimeout.
func WithContainerShutdownTimeout(timeout time.Duration) Option {
	return func(c *Options) {
		c.ContainerShutdownTimeout = timeout
	}
}

// WithNetworkCreateTimeout sets the timeout for creating the container network.
// The default is DefaultNetworkCreateTimeout.
func WithNetworkCreateTimeout(timeout time.Duration) Option {
	return func(c *Options) {
		c.NetworkCreateTimeout = timeout
	}
}

// WithNetworkShutdownTimeout sets the timeout for shutting down the container network.
// The default is DefaultNetworkShutdownTimeout.
func WithNetworkShutdownTimeout(timeout time.Duration) Option {
	return func(c *Options) {
		c.NetworkShutdownTimeout = timeout
	}
}

// WithLogger sets the logger.
// The default is DefaultLogger.
func WithLogger(logger testcontainers.Logging) Option {
	return func(c *Options) {
		c.Logger = logger
	}
}
