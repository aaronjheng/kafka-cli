package e2e_test

import (
	"os"
	"testing"
)

const (
	tlsTestBroker = "127.0.0.1:9095"
	tlsCACertEnv  = "TLS_TEST_CA_CERT"
)

func requireTLS(t *testing.T) {
	t.Helper()

	if os.Getenv(tlsCACertEnv) == "" {
		t.Skip("TLS_TEST_CA_CERT not set, skipping TLS E2E test")
	}
}

func tlsConfig() string {
	caCertPath := os.Getenv(tlsCACertEnv)

	return `
default_cluster: tls

clusters:
  tls:
    brokers:
      - ` + tlsTestBroker + `
    tls:
      insecure: true
      cafile: ` + caCertPath + `
`
}

func TestClusterDescribe_TLS(t *testing.T) {
	t.Parallel()
	requireTLS(t)

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, tlsConfig())

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
