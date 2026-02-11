package admin

import (
	"github.com/IBM/sarama"
)

type Admin struct {
	clusterAdmin sarama.ClusterAdmin
}

func NewAdmin(clusterAdmin sarama.ClusterAdmin) *Admin {
	return &Admin{
		clusterAdmin: clusterAdmin,
	}
}
