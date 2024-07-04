package ptah_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/pkg/errors"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
	"strings"
)

func (e *taskExecutor) createDockerService(ctx context.Context, req *t.CreateServiceReq) (*t.CreateServiceRes, error) {
	var res t.CreateServiceRes

	spec, err := e.prepareServicePayload(ctx, req.ServicePayload, req.SecretVars)
	if err != nil {
		return nil, errors.Wrapf(err, "create docker service")
	}

	response, err := e.docker.ServiceCreate(ctx, *spec, types.ServiceCreateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "create docker service")
	}

	res.Docker.ID = response.ID

	return &res, nil
}

func (e *taskExecutor) updateDockerService(ctx context.Context, req *t.UpdateServiceReq) (*t.UpdateServiceRes, error) {
	var res t.UpdateServiceRes

	spec, err := e.prepareServicePayload(ctx, req.ServicePayload, req.SecretVars)
	if err != nil {
		return nil, errors.Wrapf(err, "update docker service")
	}

	services, err := e.docker.ServiceList(ctx, types.ServiceListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", req.SwarmServiceSpec.Name),
		),
	})

	if err != nil {
		return nil, errors.Wrapf(err, "update docker service")
	}

	if len(services) > 1 {
		return nil, fmt.Errorf("multiple services with name %s found", req.SwarmServiceSpec.Name)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("service with name %s not found", req.SwarmServiceSpec.Name)
	}

	service := services[0]

	_, err = e.docker.ServiceUpdate(ctx, service.ID, service.Version, *spec, types.ServiceUpdateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "update docker service")
	}

	return &res, nil
}

func (e *taskExecutor) prepareServicePayload(ctx context.Context, servicePayload t.ServicePayload, secretVars t.SecretVars) (*swarm.ServiceSpec, error) {
	spec := servicePayload.SwarmServiceSpec

	if secretVars.ConfigName != "" {
		newVars := make(map[string]string)

		for key, value := range secretVars.Values {
			newVars[key] = value
		}

		if secretVars.PreserveFromConfig != "" {
			preserveConfig, err := e.getConfigByName(ctx, secretVars.PreserveFromConfig)
			if err != nil {
				return nil, errors.Wrapf(err, "get config by name %s", secretVars.PreserveFromConfig)
			}

			preservedVars := make(map[string]string)
			err = json.Unmarshal(preserveConfig.Spec.Data, &preservedVars)
			if err != nil {
				return nil, errors.Wrapf(err, "unmarshal config %s", secretVars.PreserveFromConfig)
			}

			for _, key := range secretVars.Preserve {
				if value, ok := preservedVars[key]; ok {
					newVars[key] = value
				}
			}
		}

		newVarsJson, err := json.Marshal(newVars)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal vars")
		}

		_, err = e.docker.ConfigCreate(ctx, swarm.ConfigSpec{
			Annotations: swarm.Annotations{
				Name:   secretVars.ConfigName,
				Labels: secretVars.ConfigLabels,
			},
			Data: newVarsJson,
		})

		if err != nil {
			return nil, errors.Wrapf(err, "create config %s", secretVars.ConfigName)
		}

		for key, value := range newVars {
			spec.TaskTemplate.ContainerSpec.Env = append(spec.TaskTemplate.ContainerSpec.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	for _, config := range spec.TaskTemplate.ContainerSpec.Configs {
		cfg, err := e.getConfigByName(ctx, config.ConfigName)
		if err != nil {
			return nil, errors.Wrapf(err, "get config by name %s", config.ConfigName)
		}

		config.ConfigID = cfg.ID
	}

	if servicePayload.ReleaseCommand.Command != "" {
		image, _, err := e.docker.ImageInspectWithRaw(ctx, spec.TaskTemplate.ContainerSpec.Image)
		if err != nil {
			return nil, errors.Wrapf(err, "get image %s", spec.TaskTemplate.ContainerSpec.Image)
		}

		entrypoint := strings.Join(image.Config.Entrypoint, " ")
		command := strings.Join(image.Config.Cmd, " ")

		originalEntrypoint := entrypoint + " " + command

		script := []string{
			"#!/bin/sh",
			"set -e",
			"echo 'Starting release command'",
			servicePayload.ReleaseCommand.Command,
			"echo 'Release command finished'",
			"echo 'Starting original entrypoint'",
			originalEntrypoint,
		}

		config := swarm.ConfigSpec{
			Annotations: swarm.Annotations{
				Name:   servicePayload.ReleaseCommand.ConfigName,
				Labels: servicePayload.ReleaseCommand.ConfigLabels,
			},
			Data: []byte(strings.Join(script, "\n")),
		}

		config.Labels["kind"] = "entrypoint"

		resp, err := e.docker.ConfigCreate(ctx, config)
		if err != nil {
			return nil, errors.Wrapf(err, "create config %s", servicePayload.ReleaseCommand.ConfigName)
		}

		spec.TaskTemplate.ContainerSpec.Configs = append(spec.TaskTemplate.ContainerSpec.Configs, &swarm.ConfigReference{
			File: &swarm.ConfigReferenceFileTarget{
				Name: "/ptah/entrypoint.sh",
				UID:  "0",
				GID:  "0",
				Mode: 0644,
			},
			ConfigID:   resp.ID,
			ConfigName: servicePayload.ReleaseCommand.ConfigName,
		})

		spec.TaskTemplate.ContainerSpec.Command = []string{
			"sh", "/ptah/entrypoint.sh",
		}
	}

	for _, secret := range spec.TaskTemplate.ContainerSpec.Secrets {
		secrets, err := e.docker.SecretList(ctx, types.SecretListOptions{
			Filters: filters.NewArgs(
				filters.Arg("name", secret.SecretName),
			),
		})

		if err != nil {
			return nil, errors.Wrapf(err, "get secret by name %s", secret.SecretName)
		}

		if len(secrets) > 1 {
			return nil, fmt.Errorf("multiple secrets with name %s found", secret.SecretName)
		}

		if len(secrets) == 0 {
			return nil, fmt.Errorf("secret with name %s not found", secret.SecretName)
		}

		secret.SecretID = secrets[0].ID
	}

	return &spec, nil
}

func (e *taskExecutor) deleteDockerService(ctx context.Context, req *t.DeleteServiceReq) (*t.DeleteServiceRes, error) {
	var res t.DeleteServiceRes

	services, err := e.docker.ServiceList(ctx, types.ServiceListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", req.ServiceName),
		),
	})

	if err != nil {
		return nil, errors.Wrapf(err, "delete docker service")
	}

	if len(services) == 0 {
		// TODO: return warnings if the service has not been found
		//return nil, fmt.Errorf("service with name %s not found", req.ServiceName)
		return nil, nil
	}

	err = e.docker.ServiceRemove(ctx, services[0].ID)
	if err != nil {
		return nil, errors.Wrapf(err, "delete docker service")
	}

	return &res, nil
}
