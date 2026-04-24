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

	cfgPath := filepath.Join(k.configDir, "config.yaml")

	err := os.WriteFile(cfgPath, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
}

func (k *KafkaCLI) Run(ctx context.Context, args ...string) (string, error) {
	allArgs := append([]string{"-f", filepath.Join(k.configDir, "config.yaml")}, args...)

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
	allArgs := append([]string{"-f", filepath.Join(k.configDir, "config.yaml")}, args...)

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
		"-f", filepath.Join(cli.configDir, "config.yaml"),
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
