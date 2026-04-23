package e2e_test

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestTopicCRUD_SASL(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: sasl

clusters:
  sasl:
    brokers:
      - 127.0.0.1:9094
    sasl:
      mechanism: PLAIN
      username: admin
      password: admin-secret
`)

	WaitForKafka(t, cli)

	topic := UniqueTopicName(t)

	_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "3", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic with SASL failed: %v", err)
	}

	output, err := cli.Run(t.Context(), "topic", "list")
	if err != nil {
		t.Fatalf("list topics with SASL failed: %v", err)
	}

	if !TopicExistsInOutput(output, topic) {
		t.Errorf("topic %q not found in SASL list output", topic)
	}

	describeOutput, err := cli.Run(t.Context(), "topic", "describe", topic)
	if err != nil {
		t.Fatalf("describe topic with SASL failed: %v", err)
	}

	partitions := ExtractPartitionCount(describeOutput)
	if partitions != "3" {
		t.Errorf("expected 3 partitions with SASL, got %q", partitions)
	}

	_, err = cli.Run(t.Context(), "topic", "delete", topic)
	if err != nil {
		t.Fatalf("delete topic with SASL failed: %v", err)
	}
}

func TestClusterDescribe_SASL(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: sasl

clusters:
  sasl:
    brokers:
      - 127.0.0.1:9094
    sasl:
      mechanism: PLAIN
      username: admin
      password: admin-secret
`)

	WaitForKafka(t, cli)

	output, err := cli.Run(t.Context(), "cluster", "describe")
	if err != nil {
		t.Fatalf("cluster describe with SASL failed: %v", err)
	}

	if !StringsContains(output, "Controller ID:") {
		t.Errorf("expected 'Controller ID:' in SASL cluster describe output, got: %s", output)
	}
}

func TestGroupList_SASL(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	cli.WriteConfig(t, `
default_cluster: sasl

clusters:
  sasl:
    brokers:
      - 127.0.0.1:9094
    sasl:
      mechanism: PLAIN
      username: admin
      password: admin-secret
`)

	WaitForKafka(t, cli)

	_, err := cli.Run(t.Context(), "group", "list")
	if err != nil {
		t.Fatalf("list consumer groups with SASL failed: %v", err)
	}
}

func TestTopicCRUD_TLS(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	caPath := filepath.Join(".", "certs", "ca.crt")
	cli.WriteConfig(t, fmt.Sprintf(`
default_cluster: tls

clusters:
  tls:
    brokers:
      - 127.0.0.1:9095
    tls:
      insecure: true
      cafile: %s
`, caPath))

	WaitForKafka(t, cli)

	topic := UniqueTopicName(t)

	_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "3", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic with TLS failed: %v", err)
	}

	output, err := cli.Run(t.Context(), "topic", "list")
	if err != nil {
		t.Fatalf("list topics with TLS failed: %v", err)
	}

	if !TopicExistsInOutput(output, topic) {
		t.Errorf("topic %q not found in TLS list output", topic)
	}

	_, err = cli.Run(t.Context(), "topic", "delete", topic)
	if err != nil {
		t.Fatalf("delete topic with TLS failed: %v", err)
	}
}

func TestClusterDescribe_TLS(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	caPath := filepath.Join(".", "certs", "ca.crt")
	cli.WriteConfig(t, fmt.Sprintf(`
default_cluster: tls

clusters:
  tls:
    brokers:
      - 127.0.0.1:9095
    tls:
      insecure: true
      cafile: %s
`, caPath))

	WaitForKafka(t, cli)

	output, err := cli.Run(t.Context(), "cluster", "describe")
	if err != nil {
		t.Fatalf("cluster describe with TLS failed: %v", err)
	}

	if !StringsContains(output, "Controller ID:") {
		t.Errorf("expected 'Controller ID:' in TLS cluster describe output, got: %s", output)
	}
}

func TestTopicCRUD_SASL_SSL(t *testing.T) {
	t.Parallel()

	cli := NewKafkaCLI(t)
	caPath := filepath.Join(".", "certs", "ca.crt")
	cli.WriteConfig(t, fmt.Sprintf(`
default_cluster: sasl-ssl

clusters:
  sasl-ssl:
    brokers:
      - 127.0.0.1:9096
    tls:
      insecure: true
      cafile: %s
    sasl:
      mechanism: PLAIN
      username: admin
      password: admin-secret
`, caPath))

	WaitForKafka(t, cli)

	topic := UniqueTopicName(t)

	_, err := cli.Run(t.Context(), "topic", "create", topic, "--partitions", "3", "--replication-factor", "1")
	if err != nil {
		t.Fatalf("create topic with SASL+SSL failed: %v", err)
	}

	output, err := cli.Run(t.Context(), "topic", "list")
	if err != nil {
		t.Fatalf("list topics with SASL+SSL failed: %v", err)
	}

	if !TopicExistsInOutput(output, topic) {
		t.Errorf("topic %q not found in SASL+SSL list output", topic)
	}

	_, err = cli.Run(t.Context(), "topic", "delete", topic)
	if err != nil {
		t.Fatalf("delete topic with SASL+SSL failed: %v", err)
	}
}
