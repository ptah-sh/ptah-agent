package ptah_agent

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/pkg/errors"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

var (
	ErrConfigNotFound  = fmt.Errorf("config not found")
	ErrServiceNotFound = fmt.Errorf("service not found")
	ErrSecretNotFound  = fmt.Errorf("secret not found")
)

func (e *taskExecutor) createDockerConfig(ctx context.Context, req *t.CreateConfigReq) (*t.CreateConfigRes, error) {
	var res t.CreateConfigRes

	response, err := e.docker.ConfigCreate(ctx, req.SwarmConfigSpec)
	if err != nil {
		return nil, err
	}

	res.Docker.ID = response.ID

	return &res, nil
}

func (e *taskExecutor) getConfigByName(ctx context.Context, name string) (*swarm.Config, error) {
	if name == "" {
		return nil, errors.Wrapf(ErrConfigNotFound, "config name is empty")
	}

	configs, err := e.docker.ConfigList(ctx, types.ConfigListOptions{
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
