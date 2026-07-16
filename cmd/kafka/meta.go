package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aaronjheng/kafka-cli/internal/config"
	"github.com/aaronjheng/kafka-cli/internal/kafka/admin"
)

type Meta struct {
	Config  func() (*config.Config, error)
	Cluster func() string
}

func NewMeta(cfgFilepath, cluster *string) *Meta {
	var (
		cachedCfg *config.Config
		cfgErr    error
	)

	return &Meta{
		Config: func() (*config.Config, error) {
			if cachedCfg != nil || cfgErr != nil {
				return cachedCfg, cfgErr
			}

			cachedCfg, cfgErr = config.LoadConfig(*cfgFilepath)
			if cfgErr != nil {
				cfgErr = fmt.Errorf("load config: %w", cfgErr)
			}

			return cachedCfg, cfgErr
		},
		Cluster: func() string {
			return *cluster
		},
	}
}

func withAdmin(ctx context.Context, meta *Meta, run func(*admin.Admin) error) error {
	cfg, err := meta.Config()
	if err != nil {
		return err
	}

	adminClient, closer, err := admin.NewFromConfig(cfg, meta.Cluster())
	if err != nil {
		return fmt.Errorf("provideAdmin error: %w", err)
	}

	defer func() {
		err := closer(ctx)
		if err != nil {
			slog.Error("closer error", slog.Any("error", err))
		}
	}()

	return run(adminClient)
}
