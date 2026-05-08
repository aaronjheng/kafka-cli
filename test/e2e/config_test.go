package e2e_test

import (
	"strings"
	"testing"
)

func TestConfigCat_PlainText(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9092
`)

	output, err := cli.Run(t.Context(), "config", "cat")
	if err != nil {
		t.Fatalf("config cat failed: %v", err)
	}

	if !strings.HasPrefix(output, "# ") {
		t.Errorf("expected config path comment at start, got: %s", output)
	}

	if !StringsContains(output, "default_cluster: local") {
		t.Errorf("expected 'default_cluster: local' in output, got: %s", output)
	}

	if !StringsContains(output, "127.0.0.1:9092") {
		t.Errorf("expected broker address in output, got: %s", output)
	}
}

func TestClusterFlag_NonExistentCluster(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9092
`)

	_, err := cli.Run(t.Context(), "-c", "nonexistent", "cluster", "describe")
	if err == nil {
		t.Error("expected error when using --cluster with non-existent cluster name")
	}
}

func TestClusterFlag_SelectAlternateCluster(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9092
  alternate:
    brokers:
      - 127.0.0.1:9092
`)

	WaitForKafka(t, cli)

	output, err := cli.Run(t.Context(), "-c", "alternate", "cluster", "describe")
	if err != nil {
		t.Fatalf("cluster describe with --cluster alternate failed: %v", err)
	}

	if !StringsContains(output, "Cluster ID:") {
		t.Errorf("expected 'Cluster ID:' in output, got: %s", output)
	}
}

func TestConnectionFailure_PlainText(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: local

clusters:
  local:
    brokers:
      - 127.0.0.1:9999
`)

	_, err := cli.Run(t.Context(), "cluster", "describe")
	if err == nil {
		t.Error("expected error when connecting to non-existent broker")
	}
}
