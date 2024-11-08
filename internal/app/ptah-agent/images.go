package ptah_agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/busybox"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/docker/config"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/registry"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) buildImage(ctx context.Context, req *t.BuildImageReq) (*t.BuildImageRes, error) {
	var result t.BuildImageRes

	box := busybox.New(e.docker)

	config := &busybox.Config{
		Cmd: "/ptah/bin/docker_build.sh",
		EnvVars: []string{
			fmt.Sprintf("TARGET_DIR=%s", req.WorkingDir),
			fmt.Sprintf("IMAGE_NAME=%s", req.DockerImage),
			fmt.Sprintf("DOCKERFILE_PATH=%s", req.Dockerfile),
		},
		Mounts: []mount.Mount{req.VolumeSpec},
	}

	if err := box.Start(ctx, config); err != nil {
		return nil, fmt.Errorf("start busybox: %w", err)
	}

	defer box.Stop(ctx)

	err := box.Wait(ctx)
	if err != nil {
		logs, logsErr := box.Logs(ctx)
		if logsErr != nil {
			return nil, errors.Wrap(logsErr, "get build image logs")
		}

		message := strings.Join(logs, "\n")

		return nil, errors.Wrapf(err, "%s\nwait for build image", message)
	}

	logs, err := box.Logs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get build image logs")
	}

	result.Output = logs

	return &result, nil
}

func (e *taskExecutor) buildImageWithNixpacks(ctx context.Context, req *t.BuildImageWithNixpacksReq) (*t.BuildImageWithNixpacksRes, error) {
	var result t.BuildImageWithNixpacksRes

	box := busybox.New(e.docker)

	config := &busybox.Config{
		Cmd: "/ptah/bin/nixpacks_build.sh",
		EnvVars: []string{
			fmt.Sprintf("TARGET_DIR=%s", req.WorkingDir),
			fmt.Sprintf("IMAGE_NAME=%s", req.DockerImage),
			fmt.Sprintf("NIXPACKS_FILE_PATH=%s", req.NixpacksFilePath),
		},
		Mounts: []mount.Mount{req.VolumeSpec},
	}

	if err := box.Start(ctx, config); err != nil {
		return nil, fmt.Errorf("start busybox: %w", err)
	}

	defer box.Stop(ctx)

	err := box.Wait(ctx)
	if err != nil {
		logs, logsErr := box.Logs(ctx)
		if logsErr != nil {
			return nil, errors.Wrap(logsErr, "get build image logs")
		}

		message := strings.Join(logs, "\n")

		return nil, errors.Wrapf(err, "%s\nwait for build image", message)
	}

	logs, err := box.Logs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get build image logs")
	}

	result.Output = logs

	return &result, nil
}

func (e *taskExecutor) pullImage(ctx context.Context, req *t.PullImageReq) (*t.PullImageRes, error) {
	if req.AuthConfigName != "" {
		config, err := config.GetByName(ctx, e.docker, req.AuthConfigName)
		if err != nil {
			return nil, fmt.Errorf("pull image: get config by name: %w", err)
		}

		authConfig := base64.StdEncoding.EncodeToString(config.Spec.Data)
		req.PullOptionsSpec.RegistryAuth = authConfig
	}

	outputStream, err := e.docker.ImagePull(ctx, req.Image, req.PullOptionsSpec)
	if err != nil {
		return nil, fmt.Errorf("pull image: %w", err)
	}
	defer outputStream.Close()

	decoder := json.NewDecoder(outputStream)
	var output []string

	for {
		var line struct {
			Status string `json:"status"`
			Error  string `json:"error,omitempty"`
		}

		if err := decoder.Decode(&line); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("pull image: decode json: %w", err)
		}

		if line.Error != "" {
			return nil, fmt.Errorf("pull image: docker error: %s", line.Error)
		}

		if line.Status != "" {
			output = append(output, line.Status)
		}
	}

	return &t.PullImageRes{Output: output}, nil
}

func (e *taskExecutor) pruneDockerRegistry(ctx context.Context, req *t.PruneDockerRegistryReq) (*t.PruneDockerRegistryRes, error) {
	log := Logger(ctx)

	var result t.PruneDockerRegistryRes

	tagsToKeep := make(map[string][]string)
	for _, imageRef := range req.KeepImages {
		ref, err := reference.ParseNamed(imageRef)
		if err != nil {
			return nil, fmt.Errorf("prune docker registry: %w", err)
		}

		repo := reference.Path(ref)

		repoTags, ok := tagsToKeep[repo]
		if !ok {
			repoTags = make([]string, 0)
		}

		taggedRef, ok := reference.TagNameOnly(ref).(reference.Tagged)
		if !ok {
			return nil, fmt.Errorf("prune docker registry: can not get tag from ref: %s", imageRef)
		}

		tagsToKeep[repo] = append(repoTags, taggedRef.Tag())

		log.Debug("parsed image ref", "ref", ref, "repo", repo, "tag", taggedRef.Tag())
	}

	registry := registry.New("http://registry.ptah.local:5050")

	catalog, err := registry.Catalog(ctx)
	if err != nil {
		return nil, fmt.Errorf("prune docker registry: %w", err)
	}

	for _, repo := range catalog.Repositories {
		log.Debug("processing repo", "repo", repo)

		tags, err := registry.TagsList(ctx, repo)
		if err != nil {
			return nil, fmt.Errorf("prune docker registry: %w", err)
		}

		for _, tag := range tags.Tags {
			log.Debug("processing tag", "tag", tag)

			if _, ok := tagsToKeep[repo]; !ok || !slices.Contains(tagsToKeep[repo], tag) {
				manifest, err := registry.ManifestHead(ctx, repo, tag)
				if err != nil {
					return nil, fmt.Errorf("prune docker registry: %w", err)
				}

				log.Debug("deleting image", "repo", repo, "tag", tag, "digest", manifest.Digest)

				if err := registry.ManifestDelete(ctx, repo, manifest.Digest); err != nil {
					return nil, fmt.Errorf("prune docker registry: %w", err)
				}
			}
		}
	}

	// ref, err := reference.ParseNamed("ptah/test")
	// if err != nil {
	// 	return nil, fmt.Errorf("prune docker registry: %w", err)
	// }

	// info, err := dockerRegistry.ParseRepositoryInfo(ref)
	// if err != nil {
	// 	return nil, fmt.Errorf("prune docker registry: %w", err)
	// }

	return &result, nil
}
