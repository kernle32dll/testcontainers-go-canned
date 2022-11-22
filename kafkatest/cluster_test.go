package kafkatest_test

import (
	"context"
	"testing"
	"time"

	"github.com/kernle32dll/testcontainers-go-canned/kafkatest"
	"github.com/testcontainers/testcontainers-go"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestRoundTrip_WithoutTLS(t *testing.T) {
	// Given
	t.Log("Staring new kafka cluster")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// when
	kafkaCluster, err := kafkatest.NewKafkaCluster(
		ctx,
		kafkatest.WithLogger(testcontainers.TestLogger(t)),
	)
	t.Cleanup(shutdownCluster(t, kafkaCluster))

	// then
	if err != nil {
		t.Fatal(err)
	}

	testRoundTrip(t, kafkaCluster)
}

func TestRoundTrip_WithTLS(t *testing.T) {
	// Given
	t.Log("Staring new WithTLS kafka cluster")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// when
	kafkaCluster, err := kafkatest.NewKafkaCluster(
		ctx,
		kafkatest.WithLogger(testcontainers.TestLogger(t)),
		kafkatest.WithTLS(true),
	)
	t.Cleanup(shutdownCluster(t, kafkaCluster))

	// then
	if err != nil {
		t.Fatal(err)
	}

	testRoundTrip(t, kafkaCluster)
}

func shutdownCluster(t *testing.T, kafkaCluster *kafkatest.KafkaCluster) func() {
	return func() {
		t.Helper()

		if kafkaCluster == nil {
			t.Log("Kafka cluster not started, skipping teardown")
		}

		t.Log("Tearing down kafka cluster")
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		kafkaCluster.StopCluster(ctx)
	}
}

func testRoundTrip(t *testing.T, kafkaCluster *kafkatest.KafkaCluster) {
	if err := createTopic(t, kafkaCluster, t.Name()); err != nil {
		t.Fatal(err)
	}
}

func createTopic(t *testing.T, kafkaCluster *kafkatest.KafkaCluster, topicName string) error {
	t.Helper()

	t.Log("Initializing kafka admin client")
	seed, err := kafkaCluster.GetKafkaHostAndPort(context.Background())
	if err != nil {
		return err
	}

	kadmClient, err := kadm.NewOptClient(
		kgo.SeedBrokers(seed),
		kgo.DialTLSConfig(kafkaCluster.ClientTLSConfig()),
	)
	if err != nil {
		return err
	}

	t.Log("Trying to create topic....")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if _, err := kadmClient.CreateTopics(ctx, 3, 1, nil, topicName); err != nil {
		return err
	}

	return nil
}
