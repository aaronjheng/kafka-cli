package admin

import (
	"github.com/IBM/sarama"
)

type Admin struct {
	clusterAdmin sarama.ClusterAdmin
	client       sarama.Client
}

func NewAdmin(clusterAdmin sarama.ClusterAdmin, client sarama.Client) *Admin {
	return &Admin{
		clusterAdmin: clusterAdmin,
		client:       client,
	}
}
