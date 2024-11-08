package config

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/pkg/errors"

	dockerClient "github.com/docker/docker/client"
)

var (
	ErrConfigNotFound = fmt.Errorf("config not found")
)

func GetByName(ctx context.Context, docker *dockerClient.Client, name string) (*swarm.Config, error) {
	if name == "" {
		return nil, errors.Wrapf(ErrConfigNotFound, "config name is empty")
	}

	configs, err := docker.ConfigList(ctx, types.ConfigListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", name),
		),
	})

	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.Spec.Name == name {
			return &config, nil
		}
	}

	return nil, errors.Wrapf(ErrConfigNotFound, "config with name %s not found", name)
}
