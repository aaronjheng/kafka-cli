package admin

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/aaronjheng/kafka-cli/internal/config"
	"github.com/aaronjheng/kafka-cli/internal/kafka"
)

type Admin struct {
	client       sarama.Client
	clusterAdmin sarama.ClusterAdmin
}

func NewAdmin(client sarama.Client, clusterAdmin sarama.ClusterAdmin) *Admin {
	return &Admin{
		client:       client,
		clusterAdmin: clusterAdmin,
	}
}

func NewFromConfig(cfg *config.Config, clusterName string) (*Admin, func(context.Context) error, error) {
	clusterCfg, err := cfg.Cluster(clusterName)
	if err != nil {
		return nil, nil, fmt.Errorf("cfg.Cluster error: %w", err)
	}

	cluster, err := kafka.New(clusterCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("kafka.New error: %w", err)
	}

	clusterAdmin, err := sarama.NewClusterAdminFromClient(cluster)
	if err != nil {
		return nil, nil, fmt.Errorf("newClusterAdmin error: %w", err)
	}

	return NewAdmin(cluster, clusterAdmin), func(_ context.Context) error {
		return clusterAdmin.Close()
	}, nil
}
