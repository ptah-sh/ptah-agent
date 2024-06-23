package ptah_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) createDockerService(ctx context.Context, req *t.CreateServiceReq) (*t.CreateServiceRes, error) {
	var res t.CreateServiceRes

	if req.Payload.AuthConfigName != "" {
		// read auth data from the docker config
	}

	if req.Payload.SecretVarsConfigName != "" {
		configs, err := e.docker.ConfigList(ctx, types.ConfigListOptions{
			Filters: filters.NewArgs(
				filters.Arg("name", req.Payload.SecretVarsConfigName),
			),
		})

		if err != nil {
			return nil, err
		}

		if len(configs) > 1 {
			return nil, fmt.Errorf("multiple configs with name %s found", req.Payload.SecretVarsConfigName)
		}

		if len(configs) == 0 {
			return nil, fmt.Errorf("config with name %s not found", req.Payload.SecretVarsConfigName)
		}

		var secretVars []string
		err = json.Unmarshal(configs[0].Spec.Data, &secretVars)
		if err != nil {
			return nil, err
		}

		for _, secretVar := range secretVars {
			req.Payload.SwarmServiceSpec.TaskTemplate.ContainerSpec.Env = append(req.Payload.SwarmServiceSpec.TaskTemplate.ContainerSpec.Env, secretVar)
		}
	}

	for _, config := range req.Payload.SwarmServiceSpec.TaskTemplate.ContainerSpec.Configs {
		configs, err := e.docker.ConfigList(ctx, types.ConfigListOptions{
			Filters: filters.NewArgs(
				filters.Arg("name", config.ConfigName),
			),
		})

		if err != nil {
			return nil, err
		}

		if len(configs) > 1 {
			return nil, fmt.Errorf("multiple configs with name %s found", config.ConfigName)
		}

		if len(configs) == 0 {
			return nil, fmt.Errorf("config with name %s not found", config.ConfigName)
		}

		config.ConfigID = configs[0].ID
	}

	for _, secret := range req.Payload.SwarmServiceSpec.TaskTemplate.ContainerSpec.Secrets {
		secrets, err := e.docker.SecretList(ctx, types.SecretListOptions{
			Filters: filters.NewArgs(
				filters.Arg("name", secret.SecretName),
			),
		})

		if err != nil {
			return nil, err
		}

		if len(secrets) > 1 {
			return nil, fmt.Errorf("multiple secrets with name %s found", secret.SecretName)
		}

		if len(secrets) == 0 {
			return nil, fmt.Errorf("secret with name %s not found", secret.SecretName)
		}

		secret.SecretID = secrets[0].ID
	}

	response, err := e.docker.ServiceCreate(ctx, req.Payload.SwarmServiceSpec, types.ServiceCreateOptions{})
	if err != nil {
		return nil, err
	}

	res.Docker.ID = response.ID

	return &res, nil
}
