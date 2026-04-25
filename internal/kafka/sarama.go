package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
)

type saramaSCRAMClient struct {
	hashGen scram.HashGeneratorFcn
	conv    *scram.ClientConversation
}

func newSaramaSCRAMClient(hashGen scram.HashGeneratorFcn) *saramaSCRAMClient {
	return &saramaSCRAMClient{hashGen: hashGen, conv: nil}
}

func (c *saramaSCRAMClient) Begin(userName, password, authzID string) error {
	client, err := c.hashGen.NewClient(userName, password, authzID)
	if err != nil {
		return fmt.Errorf("scram.NewClient error: %w", err)
	}

	c.conv = client.NewConversation()

	return nil
}

func (c *saramaSCRAMClient) Step(challenge string) (string, error) {
	resp, err := c.conv.Step(challenge)
	if err != nil {
		return resp, fmt.Errorf("scram.Step error: %w", err)
	}

	return resp, nil
}

func (c *saramaSCRAMClient) Done() bool {
	return c.conv.Done()
}

var _ sarama.SCRAMClient = (*saramaSCRAMClient)(nil)
