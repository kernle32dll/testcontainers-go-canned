![test](https://github.com/kernle32dll/testcontainers-go-canned/workflows/test/badge.svg)
[![GoDoc](https://godoc.org/github.com/kernle32dll/testcontainers-go-canned?status.svg)](http://godoc.org/github.com/kernle32dll/testcontainers-go-canned)
[![Go Report Card](https://goreportcard.com/badge/github.com/kernle32dll/testcontainers-go-canned)](https://goreportcard.com/report/github.com/kernle32dll/testcontainers-go-canned)
[![codecov](https://codecov.io/gh/kernle32dll/testcontainers-go-canned/branch/master/graph/badge.svg)](https://codecov.io/gh/kernle32dll/testcontainers-go-canned)

# testcontainers-go-canned

testcontainers-go-canned is a collection templates for
[testcontainers-go](https://github.com/testcontainers/testcontainers-go).

Download:

```
go get github.com/kernle32dll/testcontainers-go-canned@latest
```

Detailed documentation can be found on [GoDoc](https://godoc.org/github.com/kernle32dll/testcontainers-go-canned).

## Compatibility

testcontainers-go-canned is automatically tested against the following:

- Go 1.17.X, 1.18.X and 1.19.X

## Getting started

### Kafka

Simply initialize a new cluster via...

```go
import "github.com/kernle32dll/testcontainers-go-canned/kafkatest"

[...]

kafkaCluster, err := kafkatest.NewKafkaCluster(ctx)
```

...and if done, call...

```go
kafkaCluster.StopCluster(ctx)
```

For some inspiration, look at the [cluster_test.go](./kafkatest/cluster_test.go) file, which actually tests
integration with [franz-go](https://github.com/twmb/franz-go).

#### Configuration

There are many `kafkatest.With[...]` option functions, you can use for vararg configuration of the cluster. E.g.:

```go
kafkaCluster, err := kafkatest.NewKafkaCluster(
    ctx,
    kafkatest.WithTLS(true),
)
```

#### TLS

Take special note of the `kafkatest.WithTLS(true)` option, which configures the cluster for full TLS.
You can then use `kafkaCluster.ClientTLSConfig()` function to receive an `tls.Config`, you can plug into your
kafka client like so for example ([kgo is franz-go](https://github.com/twmb/franz-go)):

```go
seed, err := kafkaCluster.GetKafkaHostAndPort(context.Background())
[...]

client, err := kgo.NewClient(
    kgo.SeedBrokers(seed),
    kgo.DialTLSConfig(kafkaCluster.ClientTLSConfig()),
)
```