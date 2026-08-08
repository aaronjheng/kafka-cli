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
	"sync"
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

	_, err := os.Stat(binary)
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

	cmd := exec.CommandContext(ctx, k.binary, allArgs...)

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

	cmd := exec.CommandContext(ctx, k.binary, allArgs...)

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

	consumeCtx, cancelConsume := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancelConsume()

	consumerCmdArgs := make([]string, 0, 5+len(consumeArgs))
	consumerCmdArgs = append(consumerCmdArgs,
		"-f", filepath.Join(cli.configDir, "kafka.yaml"),
		"topic", "consume", topic,
	)
	consumerCmdArgs = append(consumerCmdArgs, consumeArgs...)

	consumerCmd := exec.CommandContext(consumeCtx, cli.binary, consumerCmdArgs...)

	var consumeStdout lockedBuffer

	var consumeStderr lockedBuffer

	consumerCmd.Stdout = &consumeStdout
	consumerCmd.Stderr = &consumeStderr

	err := consumerCmd.Start()
	if err != nil {
		t.Fatalf("start consumer failed: %v", err)
	}

	// Give the consumer a moment to establish its connection and offsets.
	time.Sleep(1 * time.Second)

	_, err = cli.RunWithStdin(t.Context(), messages, "topic", "produce", topic)
	if err != nil {
		cancelConsume()

		_ = consumerCmd.Wait()

		t.Fatalf("produce messages failed: %v", err)
	}

	if !waitForConsumedMessages(&consumeStdout, expectedMessages) {
		cancelConsume()

		_ = consumerCmd.Wait()

		t.Fatalf(
			"timed out waiting for consumed messages, got stdout=%q stderr=%q",
			consumeStdout.String(),
			consumeStderr.String(),
		)
	}

	cancelConsume()

	_ = consumerCmd.Wait()
}

// waitForConsumedMessages polls the consumer output until every expected
// message has been consumed or the deadline passes.
func waitForConsumedMessages(output *lockedBuffer, expectedMessages []string) bool {
	deadline := time.Now().Add(15 * time.Second)

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for !containsAll(output.String(), expectedMessages) {
		if time.Now().After(deadline) {
			return false
		}

		<-ticker.C
	}

	return true
}

// lockedBuffer is a goroutine-safe bytes.Buffer wrapper so the consumer
// subprocess can write to it while the test polls its contents.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	//nolint:wrapcheck // implements io.Writer; bytes.Buffer never returns an error
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.b.String()
}

func containsAll(s string, substrs []string) bool {
	for _, substr := range substrs {
		if !strings.Contains(s, substr) {
			return false
		}
	}

	return true
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
		expected:  len(messages),
		mu:        sync.Mutex{},
		consumed:  0,
		done:      make(chan struct{}),
		closeOnce: sync.Once{},
	}

	go consumeGroupErrors(ctx, client)
	go consumeGroupLoop(ctx, client, topic, handler)

	select {
	case <-handler.done:
	case <-ctx.Done():
		t.Fatal("timed out waiting for consumer group to consume messages")
	}
}

func AssertGroupLagForTopic(t *testing.T, cli *KafkaCLI, group string, topic string) {
	t.Helper()

	output, err := cli.Run(t.Context(), "group", "lag", group, "--topic", topic)
	if err != nil {
		t.Fatalf("group lag with topic filter failed: %v", err)
	}

	if !strings.Contains(output, group) {
		t.Errorf("expected group %q in group lag output, got: %s", group, output)
	}

	if !strings.Contains(output, topic) {
		t.Errorf("expected topic %q in group lag output, got: %s", topic, output)
	}

	if !strings.Contains(output, "Lag") {
		t.Errorf("expected 'Lag' in group lag output, got: %s", output)
	}

	if !strings.Contains(output, "Partitions") {
		t.Errorf("expected 'Partitions' in group lag output, got: %s", output)
	}

	if !strings.Contains(output, "Partition") {
		t.Errorf("expected 'Partition' in group lag output, got: %s", output)
	}
}

func AssertGroupOffsetsForTopic(t *testing.T, cli *KafkaCLI, group string, topic string) {
	t.Helper()

	output, err := cli.Run(t.Context(), "group", "offsets", group, "--topic", topic)
	if err != nil {
		t.Fatalf("group offsets with topic filter failed: %v", err)
	}

	if !strings.Contains(output, group) {
		t.Errorf("expected group %q in group offsets output, got: %s", group, output)
	}

	if !strings.Contains(output, topic) {
		t.Errorf("expected topic %q in group offsets output, got: %s", topic, output)
	}

	if !strings.Contains(output, "Partition") {
		t.Errorf("expected 'Partition' in group offsets output, got: %s", output)
	}

	if !strings.Contains(output, "Current Offset") {
		t.Errorf("expected 'Current Offset' in group offsets output, got: %s", output)
	}

	if !strings.Contains(output, "Log End Offset") {
		t.Errorf("expected 'Log End Offset' in group offsets output, got: %s", output)
	}

	if !strings.Contains(output, "Lag") {
		t.Errorf("expected 'Lag' in group offsets output, got: %s", output)
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

	mu        sync.Mutex
	consumed  int
	done      chan struct{}
	closeOnce sync.Once
}

func (h *consumeGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumeGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h *consumeGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		sess.MarkMessage(msg, "")
		sess.Commit()

		h.mu.Lock()
		h.consumed++
		reached := h.consumed >= h.expected
		h.mu.Unlock()

		if reached {
			h.closeOnce.Do(func() { close(h.done) })

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

var waitForKafkaOnce sync.Once

var errWaitForKafka error

// WaitForKafka blocks until the shared Kafka broker is ready. The broker is
// polled only once across all parallel tests; subsequent calls reuse the result.
func WaitForKafka(t *testing.T, cli *KafkaCLI) {
	t.Helper()

	waitForKafkaOnce.Do(func() {
		errWaitForKafka = waitUntilKafkaReady(cli)
	})

	if errWaitForKafka != nil {
		t.Fatalf("kafka did not become ready: %v", errWaitForKafka)
	}
}

func waitUntilKafkaReady(cli *KafkaCLI) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for kafka to be ready: %w", ctx.Err())
		case <-time.After(3 * time.Second):
		}

		_, err := cli.Run(ctx, "cluster", "describe")
		if err == nil {
			return nil
		}
	}
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func TopicExistsInOutput(output, topic string) bool {
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		cleaned := strings.Trim(stripANSI(line), " │")

		fields := strings.Fields(cleaned)
		if len(fields) > 0 && fields[0] == topic {
			return true
		}
	}

	return false
}

// waitForTopicListState polls `topic list` until every topic in topics is
// present (want=true) or absent (want=false), or the deadline elapses.
//
// Kafka (KRaft in particular) may briefly serve a MetadataResponse that omits
// a just-created or just-deleted topic, so topic CRUD tests must poll instead
// of asserting on a single list call.
func waitForTopicListState(t *testing.T, cli *KafkaCLI, want bool, topics ...string) {
	t.Helper()

	deadline := time.Now().Add(15 * time.Second)

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var lastOutput string

	for {
		output, err := cli.Run(t.Context(), "topic", "list")
		if err != nil {
			t.Fatalf("list topics failed: %v", err)
		}

		lastOutput = output

		matched := true

		for _, topic := range topics {
			if TopicExistsInOutput(output, topic) != want {
				matched = false

				break
			}
		}

		if matched {
			return
		}

		if time.Now().After(deadline) {
			break
		}

		<-ticker.C
	}

	t.Fatalf("timed out waiting for topics %v to be present=%v in list output: %s", topics, want, lastOutput)
}

func ExtractPartitionCount(output string) string {
	re := regexp.MustCompile(`Partitions:\s*(\d+)`)

	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}

	return ""
}
