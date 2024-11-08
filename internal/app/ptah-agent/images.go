package ptah_agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/busybox"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/docker/config"
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
