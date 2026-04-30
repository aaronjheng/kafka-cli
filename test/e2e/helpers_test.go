package e2e_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

const (
	kafkaCLIPathEnv = "KAFKA_CLI_BINARY"
	defaultTimeout  = 30 * time.Second
)

type KafkaCLI struct {
	binary    string
	configDir string
}

func NewKafkaCLI(t *testing.T) *KafkaCLI {
	t.Helper()

	binary := os.Getenv(kafkaCLIPathEnv)
	if binary == "" {
		binary = filepath.Join("..", "..", "kafka")
	}

	_, err := os.Stat(binary) //nolint:gosec // test-only: path from env var
	if os.IsNotExist(err) {
		t.Fatalf("kafka CLI binary not found at %s (set %s env var)", binary, kafkaCLIPathEnv)
	}

	return &KafkaCLI{
		binary:    binary,
		configDir: t.TempDir(),
	}
}

func (k *KafkaCLI) WriteConfig(t *testing.T, content string) {
	t.Helper()

	cfgPath := filepath.Join(k.configDir, "kafka.yaml")

	err := os.WriteFile(cfgPath, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
}

func (k *KafkaCLI) Run(ctx context.Context, args ...string) (string, error) {
	allArgs := append([]string{"-f", filepath.Join(k.configDir, "kafka.yaml")}, args...)

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, k.binary, allArgs...) //nolint:gosec // test-only: binary execution

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%w: stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}

	return stdout.String(), nil
}

func (k *KafkaCLI) RunWithStdin(ctx context.Context, stdin string, args ...string) (string, error) {
	allArgs := append([]string{"-f", filepath.Join(k.configDir, "kafka.yaml")}, args...)

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, k.binary, allArgs...) //nolint:gosec // test-only: binary execution

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(stdin)

	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%w: stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}

	return stdout.String(), nil
}

func ProduceAndAssertConsumed(
	t *testing.T,
	cli *KafkaCLI,
	topic string,
	expectedMessages []string,
	consumeArgs ...string,
) {
	t.Helper()

	messages := strings.Join(expectedMessages, "\n") + "\n"

	consumeCtx, cancelConsume := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancelConsume()

	consumerCmdArgs := make([]string, 0, 5+len(consumeArgs))
	consumerCmdArgs = append(consumerCmdArgs,
		"-f", filepath.Join(cli.configDir, "kafka.yaml"),
		"topic", "consume", topic,
	)

	consumerCmdArgs = append(consumerCmdArgs, consumeArgs...)

	//nolint:gosec // test-only: controlled args
	consumerCmd := exec.CommandContext(consumeCtx, cli.binary, consumerCmdArgs...)

	var consumeStdout bytes.Buffer

	var consumeStderr bytes.Buffer

	consumerCmd.Stdout = &consumeStdout
	consumerCmd.Stderr = &consumeStderr

	err := consumerCmd.Start()
	if err != nil {
		t.Fatalf("start consumer failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	_, err = cli.RunWithStdin(t.Context(), messages, "topic", "produce", topic)
	if err != nil {
		t.Fatalf("produce messages failed: %v", err)
	}

	time.Sleep(2 * time.Second)

	cancelConsume()

	_ = consumerCmd.Wait()

	output := consumeStdout.String()
	for _, message := range expectedMessages {
		if !StringsContains(output, message) {
			t.Fatalf(
				"expected consumed message %q, got stdout=%q stderr=%q",
				message,
				output,
				consumeStderr.String(),
			)
		}
	}
}

func UniqueTopicName(t *testing.T) string {
	t.Helper()

	return fmt.Sprintf("e2e-test-%s-%d", strings.ReplaceAll(t.Name(), "/", "-"), time.Now().UnixMilli())
}

func UniqueGroupName(t *testing.T) string {
	t.Helper()

	return fmt.Sprintf("e2e-group-%s-%d", strings.ReplaceAll(t.Name(), "/", "-"), time.Now().UnixMilli())
}

func ProduceAndConsumeWithGroup(
	t *testing.T,
	cli *KafkaCLI,
	topic string,
	group string,
	messages []string,
) {
	t.Helper()

	msgs := strings.Join(messages, "\n") + "\n"

	_, err := cli.RunWithStdin(t.Context(), msgs, "topic", "produce", topic)
	if err != nil {
		t.Fatalf("produce messages failed: %v", err)
	}

	brokers := brokersFromConfig(t, cli)

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		t.Fatalf("create consumer group failed: %v", err)
	}

	defer func() {
		_ = client.Close()
	}()

	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()

	handler := &consumeGroupHandler{
		expected: len(messages),
		consumed: 0,
		done:     make(chan struct{}),
	}

	go consumeGroupErrors(ctx, client)
	go consumeGroupLoop(ctx, client, topic, handler)

	select {
	case <-handler.done:
	case <-ctx.Done():
		t.Fatal("timed out waiting for consumer group to consume messages")
	}
}

func consumeGroupErrors(ctx context.Context, client sarama.ConsumerGroup) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-client.Errors():
		}
	}
}

func consumeGroupLoop(ctx context.Context, client sarama.ConsumerGroup, topic string, handler *consumeGroupHandler) {
	for {
		err := client.Consume(ctx, []string{topic}, handler)
		if err != nil {
			return
		}

		if ctx.Err() != nil {
			return
		}
	}
}

type consumeGroupHandler struct {
	expected int
	consumed int
	done     chan struct{}
}

func (h *consumeGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumeGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h *consumeGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		sess.MarkMessage(msg, "")
		sess.Commit()

		h.consumed++

		if h.consumed >= h.expected {
			close(h.done)

			return nil
		}
	}

	return nil
}

func brokersFromConfig(t *testing.T, cli *KafkaCLI) []string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(cli.configDir, "kafka.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if broker, ok := strings.CutPrefix(trimmed, "- "); ok && strings.Contains(broker, ":") {
			return []string{broker}
		}
	}

	t.Fatal("no brokers found in config")

	return nil
}

func WaitForKafka(t *testing.T, cli *KafkaCLI) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("timed out waiting for kafka to be ready")
		case <-time.After(3 * time.Second):
		}

		_, err := cli.Run(ctx, "cluster", "describe")
		if err == nil {
			return
		}
	}
}

func TopicExistsInOutput(output, topic string) bool {
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		cleaned := strings.Trim(line, " │")

		fields := strings.Fields(cleaned)
		if len(fields) > 0 && fields[0] == topic {
			return true
		}
	}

	return false
}

func ExtractPartitionCount(output string) string {
	re := regexp.MustCompile(`Partitions:\s*(\d+)`)

	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}

	return ""
}

func StringsContains(s, substr string) bool {
	return strings.Contains(s, substr)
}
