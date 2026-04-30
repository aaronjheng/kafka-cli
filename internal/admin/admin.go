package admin

import (
	"github.com/IBM/sarama"
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
