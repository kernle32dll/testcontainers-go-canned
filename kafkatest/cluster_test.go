package kafkatest_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
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
			return
		}

		t.Log("Tearing down kafka cluster")
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		kafkaCluster.StopCluster(ctx)
	}
}

func testRoundTrip(t *testing.T, kafkaCluster *kafkatest.KafkaCluster) {
	t.Log("Initializing kafka client")
	seed, err := kafkaCluster.GetKafkaHostAndPort(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	topicName := t.Name()
	kgoClient, err := kgo.NewClient(
		kgo.SeedBrokers(seed),
		kgo.DialTLSConfig(kafkaCluster.ClientTLSConfig()),
		kgo.ConsumeTopics(topicName),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Create topic
	if err := createTopic(t, kgoClient, topicName); err != nil {
		t.Fatal(err)
	}

	// Send the test message
	sent, err := sendTestMessage(t, kgoClient, topicName)
	if err != nil {
		t.Fatal(err)
	}

	received, err := receiveMessage(t, kgoClient, topicName)
	if err != nil {
		t.Fatal(err)
	}

	if received != sent {
		t.Errorf("Received unexpected message content, wanted %q, got %q", sent, received)
	}
}

func createTopic(t *testing.T, kgoClient *kgo.Client, topicName string) error {
	t.Helper()

	t.Log("Initializing kafka admin client")
	kadmClient := kadm.NewClient(kgoClient)

	t.Log("Trying to create topic....")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if _, err := kadmClient.CreateTopics(ctx, 3, 1, nil, topicName); err != nil {
		return err
	}

	return nil
}

func sendTestMessage(t *testing.T, kgoClient *kgo.Client, topicName string) (string, error) {
	t.Helper()

	t.Log("Sending message")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	msg := uuid.NewString()
	record := &kgo.Record{Topic: topicName, Value: []byte(msg)}
	if err := kgoClient.ProduceSync(ctx, record).FirstErr(); err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return string(record.Value), nil
}

func receiveMessage(t *testing.T, kgoClient *kgo.Client, topicName string) (string, error) {
	t.Helper()

	t.Log("Receive message")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fetches := kgoClient.PollFetches(ctx)
	if err := fetches.Err(); err != nil {
		return "", fmt.Errorf("failed to receive messages: %w", err)
	}

	if county := len(fetches); county != 1 {
		return "", fmt.Errorf("received unexpected message count %d", county)
	}

	msg := fetches.Records()[0]
	if msg.Topic != topicName {
		return "", fmt.Errorf("received unexpected topic message for %q", msg.Topic)
	}

	return string(msg.Value), nil
}
