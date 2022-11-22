package kafkatest_test

import (
	"io"
	"log"
	"testing"
	"time"

	"github.com/kernle32dll/testcontainers-go-canned/kafkatest"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Tests that the WithZookeeperImage option correctly applies.
func Test_Options_ZookeeperImage(t *testing.T) {
	// given
	want := "bar"
	option := kafkatest.WithZookeeperImage(want)
	options := &kafkatest.Options{ZookeeperImage: "foo"}

	// when
	option(options)

	// then
	if got := options.ZookeeperImage; got != want {
		t.Errorf("zookeeper image not correctly applied, wanted %q, got %q", want, got)
	}
}

// Tests that the WithKafkaImage option correctly applies.
func Test_Options_KafkaImage(t *testing.T) {
	// given
	want := "bar"
	option := kafkatest.WithKafkaImage(want)
	options := &kafkatest.Options{KafkaImage: "foo"}

	// when
	option(options)

	// then
	if got := options.KafkaImage; got != want {
		t.Errorf("kafka image not correctly applied, wanted %q, got %q", want, got)
	}
}

// Tests that the WithKafkaClientPort option correctly applies.
func Test_Options_KafkaClientPort(t *testing.T) {
	// given
	want := 1337
	option := kafkatest.WithKafkaClientPort(want)
	options := &kafkatest.Options{KafkaClientPort: 42}

	// when
	option(options)

	// then
	if got := options.KafkaClientPort; got != want {
		t.Errorf("kafka client port not correctly applied, wanted %d, got %d", want, got)
	}
}

// Tests that the WithNetworkPrefix option correctly applies.
func Test_Options_NetworkPrefix(t *testing.T) {
	// given
	want := "bar"
	option := kafkatest.WithNetworkPrefix(want)
	options := &kafkatest.Options{NetworkPrefix: "foo"}

	// when
	option(options)

	// then
	if got := options.NetworkPrefix; got != want {
		t.Errorf("network prefix not correctly applied, wanted %q, got %q", want, got)
	}
}

// Tests that the WithTLS option correctly applies.
func Test_Options_TLS(t *testing.T) {
	// given
	option := kafkatest.WithTLS(true)
	options := &kafkatest.Options{TLS: false}

	// when
	option(options)

	// then
	if got := options.TLS; got != true {
		t.Errorf("tls not correctly applied, wanted true, got %t", got)
	}
}

// Tests that the WithZookeeperWaitStrategy option correctly applies.
func Test_Options_ZookeeperWaitStrategy(t *testing.T) {
	// given
	want := wait.ForExposedPort()
	option := kafkatest.WithZookeeperWaitStrategy(want)
	options := &kafkatest.Options{ZookeeperWaitStrategy: wait.ForExit()}

	// when
	option(options)

	// then
	if got := options.ZookeeperWaitStrategy; got != want {
		t.Error("zookeeper wait strategy not correctly applied")
	}
}

// Tests that the WithKafkaWaitStrategy option correctly applies.
func Test_Options_KafkaWaitStrategy(t *testing.T) {
	// given
	want := wait.ForExposedPort()
	option := kafkatest.WithKafkaWaitStrategy(want)
	options := &kafkatest.Options{KafkaWaitStrategy: wait.ForExit()}

	// when
	option(options)

	// then
	if got := options.KafkaWaitStrategy; got != want {
		t.Error("kafka wait strategy not correctly applied")
	}
}

// Tests that the WithContainerStartTimeout option correctly applies.
func Test_Options_ContainerStartTimeout(t *testing.T) {
	// given
	want := time.Minute
	option := kafkatest.WithContainerStartTimeout(want)
	options := &kafkatest.Options{ContainerStartTimeout: time.Second}

	// when
	option(options)

	// then
	if got := options.ContainerStartTimeout; got != want {
		t.Errorf("container start timeout not correctly applied, wanted %s, got %s", want, got)
	}
}

// Tests that the WithContainerShutdownTimeout option correctly applies.
func Test_Options_ContainerShutdownTimeout(t *testing.T) {
	// given
	want := time.Minute
	option := kafkatest.WithContainerShutdownTimeout(want)
	options := &kafkatest.Options{ContainerShutdownTimeout: time.Second}

	// when
	option(options)

	// then
	if got := options.ContainerShutdownTimeout; got != want {
		t.Errorf("container shutdown timeout not correctly applied, wanted %s, got %s", want, got)
	}
}

// Tests that the WithNetworkCreateTimeout option correctly applies.
func Test_Options_NetworkCreateTimeout(t *testing.T) {
	// given
	want := time.Minute
	option := kafkatest.WithNetworkCreateTimeout(want)
	options := &kafkatest.Options{NetworkCreateTimeout: time.Second}

	// when
	option(options)

	// then
	if got := options.NetworkCreateTimeout; got != want {
		t.Errorf("network create timeout not correctly applied, wanted %s, got %s", want, got)
	}
}

// Tests that the WithNetworkShutdownTimeout option correctly applies.
func Test_Options_NetworkShutdownTimeout(t *testing.T) {
	// given
	want := time.Minute
	option := kafkatest.WithNetworkShutdownTimeout(want)
	options := &kafkatest.Options{NetworkShutdownTimeout: time.Second}

	// when
	option(options)

	// then
	if got := options.NetworkShutdownTimeout; got != want {
		t.Errorf("network shutdown timeout not correctly applied, wanted %s, got %s", want, got)
	}
}

// Tests that the WithLogger option correctly applies.
func Test_Options_Logger(t *testing.T) {
	// given
	want := log.New(io.Discard, "", log.LstdFlags)
	option := kafkatest.WithLogger(want)
	options := &kafkatest.Options{Logger: kafkatest.DefaultLogger}

	// when
	option(options)

	// then
	if got := options.Logger; got != want {
		t.Error("logger not correctly applied")
	}
}
