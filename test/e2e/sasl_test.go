package e2e_test

import (
	"testing"
)

const (
	saslTestBroker = "127.0.0.1:9094"
	saslMechanism  = "PLAIN"
	saslUsername   = "client"
	saslPassword   = "client-secret"
)

func saslConfig() string {
	return `
default_cluster: sasl

clusters:
  sasl:
    brokers:
      - ` + saslTestBroker + `
    sasl:
      mechanism: ` + saslMechanism + `
      username: ` + saslUsername + `
      password: ` + saslPassword + `
`
}

func TestClusterDescribe_SASL(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, saslConfig())

	WaitForKafka(t, cli)

	output, err := cli.Run(t.Context(), "cluster", "describe")
	if err != nil {
		t.Fatalf("cluster describe failed: %v", err)
	}

	if !StringsContains(output, "Controller ID:") {
		t.Errorf("expected 'Controller ID:' in output, got: %s", output)
	}

	if !StringsContains(output, "Brokers:") {
		t.Errorf("expected 'Brokers:' in output, got: %s", output)
	}

	if !StringsContains(output, "Topics:") {
		t.Errorf("expected 'Topics:' in output, got: %s", output)
	}
}
