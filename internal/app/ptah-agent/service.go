package ptah_agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/docker/config"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) launchDockerService(ctx context.Context, req *t.LaunchServiceReq) (*t.LaunchServiceRes, error) {
	ctx, log := ContextWithLoggerValues(ctx, "service", req.SwarmServiceSpec.Name)

	var res t.LaunchServiceRes

	service, err := e.getServiceByName(ctx, req.SwarmServiceSpec.Name)
	if err != nil && !errors.Is(err, ErrServiceNotFound) {
		return nil, fmt.Errorf("launch docker service: %w", err)
	}

	if service == nil {
		log.Debug("service not found, creating")

		createRes, err := e.createDockerService(ctx, (*t.CreateServiceReq)(req))
		if err != nil {
			return nil, fmt.Errorf("launch docker service (create): %w", err)
		}

		res.Action = "created"
		res.Docker.ID = createRes.Docker.ID
	} else {
		log.Debug("service found, updating")

		_, err := e.updateDockerService(ctx, (*t.UpdateServiceReq)(req))
		if err != nil {
			return nil, fmt.Errorf("launch docker service (update): %w", err)
		}

		res.Action = "updated"
		res.Docker.ID = service.ID
	}

	serviceWithRaw, _, err := e.docker.ServiceInspectWithRaw(ctx, res.Docker.ID, types.ServiceInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("launch docker service (inspect): %w", err)
	}

	service = &serviceWithRaw

	err = e.monitorServiceLaunch(ctx, service)
	if err != nil {
		return nil, fmt.Errorf("launch docker service (monitor): %w", err)
	}

	return &res, nil
}

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
	log := Logger(ctx)

	var res t.UpdateServiceRes

	spec, err := e.prepareServicePayload(ctx, req.ServicePayload, req.SecretVars)
	if err != nil {
		return nil, errors.Wrapf(err, "update docker service")
	}

	service, err := e.getServiceByName(ctx, req.SwarmServiceSpec.Name)
	if err != nil {
		return nil, err
	}

	log.Debug("update docker service", "service", service.Spec.Name, "id", service.ID, "version", service.Version)

	_, err = e.docker.ServiceUpdate(ctx, service.ID, service.Version, *spec, types.ServiceUpdateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "update docker service")
	}

	inspected, _, err := e.docker.ServiceInspectWithRaw(ctx, service.ID, types.ServiceInspectOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "inspect docker service")
	}

	log.Debug("updated docker service", "service", inspected.Spec.Name, "id", inspected.ID, "version", inspected.Version)

	return &res, nil
}

func (e *taskExecutor) prepareServicePayload(ctx context.Context, servicePayload t.ServicePayload, secretVars t.SecretVars) (*swarm.ServiceSpec, error) {
	spec := servicePayload.SwarmServiceSpec

	for key, encryptedValue := range secretVars {
		decryptedValue, err := e.decryptValue(ctx, encryptedValue)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decrypt value for %s", key)
		}

		spec.TaskTemplate.ContainerSpec.Env = append(spec.TaskTemplate.ContainerSpec.Env, fmt.Sprintf("%s=%s", key, decryptedValue))
	}

	for _, swarmConfig := range spec.TaskTemplate.ContainerSpec.Configs {
		cfg, err := config.GetByName(ctx, e.docker, swarmConfig.ConfigName)
		if err != nil {
			return nil, errors.Wrapf(err, "get config by name %s", swarmConfig.ConfigName)
		}

		swarmConfig.ConfigID = cfg.ID
	}

	for _, secret := range spec.TaskTemplate.ContainerSpec.Secrets {
		foundSecret, err := e.getSecretByName(ctx, secret.SecretName)
		if err != nil {
			return nil, errors.Wrapf(err, "get secret by name %s", secret.SecretName)
		}

		secret.SecretID = foundSecret.ID
	}

	image, _, err := e.docker.ImageInspectWithRaw(ctx, spec.TaskTemplate.ContainerSpec.Image)
	if err != nil {
		return nil, errors.Wrapf(err, "get image %s", spec.TaskTemplate.ContainerSpec.Image)
	}

	// FIXME: original entrypoint overrides custom command if both (release cmd & cmd) are set
	entrypoint := strings.Join(image.Config.Entrypoint, " ")
	command := strings.Join(image.Config.Cmd, " ")

	spec.TaskTemplate.ContainerSpec.Env = append(spec.TaskTemplate.ContainerSpec.Env, fmt.Sprintf("ENTRYPOINT=%s %s", entrypoint, command))

	if spec.TaskTemplate.ContainerSpec.Command != nil {
		entrypoint = strings.Join(spec.TaskTemplate.ContainerSpec.Command, " ")
	}

	if spec.TaskTemplate.ContainerSpec.Args != nil {
		command = strings.Join(spec.TaskTemplate.ContainerSpec.Args, " ")
	}

	if servicePayload.ReleaseCommand.Command != "" {
		script := []string{
			"#!/bin/sh",
			"set -e",
			"echo 'Starting release command'",
			servicePayload.ReleaseCommand.Command,
			"echo 'Release command finished'",
			"echo 'Starting original entrypoint'",
			fmt.Sprintf("exec %s %s", entrypoint, command),
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
