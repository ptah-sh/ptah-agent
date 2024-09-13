package ptah_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/pkg/errors"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
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

	service, err := e.getServiceByName(ctx, req.SwarmServiceSpec.Name)
	if err != nil {
		return nil, err
	}

	_, err = e.docker.ServiceUpdate(ctx, service.ID, service.Version, *spec, types.ServiceUpdateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "update docker service")
	}

	return &res, nil
}

func (e *taskExecutor) prepareServicePayload(ctx context.Context, servicePayload t.ServicePayload, secretVars t.SecretVars) (*swarm.ServiceSpec, error) {
	spec := servicePayload.SwarmServiceSpec

	if secretVars.ConfigName != "" {
		newVarsJson, err := json.Marshal(secretVars.Values)
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

		for key, value := range secretVars.Values {
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
		foundSecret, err := e.getSecretByName(ctx, secret.SecretName)
		if err != nil {
			return nil, errors.Wrapf(err, "get secret by name %s", secret.SecretName)
		}

		secret.SecretID = foundSecret.ID
	}

	return &spec, nil
}

func (e *taskExecutor) deleteDockerService(ctx context.Context, req *t.DeleteServiceReq) (*t.DeleteServiceRes, error) {
	var res t.DeleteServiceRes

	service, err := e.getServiceByName(ctx, req.ServiceName)
	if err != nil {
		return nil, err
	}

	err = e.docker.ServiceRemove(ctx, service.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "delete docker service")
	}

	return &res, nil
}

func (e *taskExecutor) getServiceByName(ctx context.Context, name string) (*swarm.Service, error) {
	if name == "" {
		return nil, errors.Wrapf(ErrServiceNotFound, "service name is empty")
	}

	services, err := e.docker.ServiceList(ctx, types.ServiceListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", name),
		),
	})

	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if service.Spec.Name == name {
			return &service, nil
		}
	}

	return nil, errors.Wrapf(ErrServiceNotFound, "service with name %s not found", name)
}
