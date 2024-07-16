package ptah_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
	"io"
	"log"
	"strings"
	"time"
)

func (e *taskExecutor) createS3Storage(ctx context.Context, req *t.CreateS3StorageReq) (*t.CreateS3StorageRes, error) {
	var res t.CreateS3StorageRes

	if req.S3StorageSpec.AccessKey == "" || req.S3StorageSpec.SecretKey == "" {
		if req.PrevConfigName == "" {
			return nil, fmt.Errorf("create s3 storage: prev config name is empty - empty credentials")
		}

		prev, err := e.getConfigByName(ctx, req.PrevConfigName)
		if err != nil {
			return nil, err
		}

		var prevSpec t.S3StorageSpec
		err = json.Unmarshal(prev.Spec.Data, &prevSpec)
		if err != nil {
			return nil, fmt.Errorf("create s3 storage: unmarshal prev config: %w", err)
		}

		req.S3StorageSpec.AccessKey = prevSpec.AccessKey
		req.S3StorageSpec.SecretKey = prevSpec.SecretKey
	}

	data, err := json.Marshal(req.S3StorageSpec)
	if err != nil {
		return nil, err
	}

	req.SwarmConfigSpec.Data = data

	config, err := e.docker.ConfigCreate(ctx, req.SwarmConfigSpec)
	if err != nil {
		return nil, err
	}

	res.Docker.ID = config.ID

	return &res, nil
}

func (e *taskExecutor) checkS3Storage(ctx context.Context, req *t.CheckS3StorageReq) (*t.CheckS3StorageRes, error) {
	err := e.uploadS3FileWithHelper(ctx, []mount.Mount{}, req.S3StorageConfigName, "/tmp/check-access.txt", ".check-access")
	if err != nil {
		return nil, err
	}

	return &t.CheckS3StorageRes{}, nil
}

func (e *taskExecutor) s3upload(ctx context.Context, req *t.S3UploadReq) (*t.S3UploadRes, error) {
	err := e.uploadS3FileWithHelper(ctx, []mount.Mount{req.VolumeSpec}, req.S3StorageConfigName, req.SrcFilePath, req.DestFilePath)
	if err != nil {
		return nil, err
	}

	return &t.S3UploadRes{}, nil
}

func (e *taskExecutor) uploadS3FileWithHelper(ctx context.Context, mounts []mount.Mount, s3StorageConfigName, srcFilePath, destFilePath string) error {
	credentialsConfig, err := e.getConfigByName(ctx, s3StorageConfigName)
	if err != nil {
		return fmt.Errorf("check s3 storage: get config: %w", err)
	}

	var s3StorageSpec t.S3StorageSpec
	err = json.Unmarshal(credentialsConfig.Spec.Data, &s3StorageSpec)
	if err != nil {
		return fmt.Errorf("check s3 storage: unmarshal config: %w", err)
	}

	uploadScript := []string{
		"set -e",
		"echo 'https://ptah.sh' > /tmp/check-access.txt",
		"s3cmd --access_key $S3_ACCESS_KEY --secret_key $S3_SECRET_KEY --host $S3_ENDPOINT --host-bucket \"%(bucket)s.$S3_ENDPOINT\" --region $S3_REGION put $SRC_FILE_PATH s3://$S3_BUCKET/$PATH_PREFIX/$DEST_FILE_PATH",
	}

	cmd := strings.Join(uploadScript, "\n")

	containerName := fmt.Sprintf("ptah-helper-%d", time.Now().Unix())

	if strings.Index(destFilePath, s3StorageSpec.PathPrefix) == 0 {
		destFilePath = destFilePath[len(s3StorageSpec.PathPrefix):]
	}

	createResponse, err := e.docker.ContainerCreate(ctx, &container.Config{
		Image: "d3fk/s3cmd",
		Entrypoint: []string{
			"sh",
			"-c",
		},
		Cmd: []string{cmd},
		Env: []string{
			fmt.Sprintf("S3_ACCESS_KEY=%s", s3StorageSpec.AccessKey),
			fmt.Sprintf("S3_SECRET_KEY=%s", s3StorageSpec.SecretKey),
			fmt.Sprintf("S3_ENDPOINT=%s", s3StorageSpec.Endpoint),
			fmt.Sprintf("S3_REGION=%s", s3StorageSpec.Region),
			fmt.Sprintf("S3_BUCKET=%s", s3StorageSpec.Bucket),
			fmt.Sprintf("PATH_PREFIX=%s", strings.Trim(s3StorageSpec.PathPrefix, "/")),
			fmt.Sprintf("SRC_FILE_PATH=%s", srcFilePath),
			fmt.Sprintf("DEST_FILE_PATH=%s", strings.Trim(destFilePath, "/")),
		},
	}, &container.HostConfig{
		Mounts: mounts,
	}, &network.NetworkingConfig{}, &v1.Platform{}, containerName)
	if err != nil {
		return fmt.Errorf("check s3 storage: create container: %w", err)
	}

	defer func(docker *dockerClient.Client, ctx context.Context, containerID string, options container.RemoveOptions) {
		err := docker.ContainerRemove(ctx, containerID, options)
		if err != nil {
			log.Printf("failed to remove container %s: %v", containerID, err)
		}
	}(e.docker, ctx, createResponse.ID, container.RemoveOptions{
		Force: true,
	})

	err = e.docker.ContainerStart(ctx, createResponse.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("check s3 storage: start container: %w", err)
	}

	waitChan, errChan := e.docker.ContainerWait(ctx, createResponse.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("check s3 storage: wait container: %w", err)
		}
	case w := <-waitChan:
		if w.Error != nil {
			return fmt.Errorf("check s3 storage: wait container: %s", w.Error.Message)
		}

		if w.StatusCode != 0 {
			logs, err := e.docker.ContainerLogs(ctx, createResponse.ID, container.LogsOptions{
				ShowStderr: true,
				ShowStdout: true,
			})
			if err != nil {
				return fmt.Errorf("check s3 storage: get logs: %w", err)
			}

			logsBytes, err := io.ReadAll(logs)
			if err != nil {
				return fmt.Errorf("check s3 storage: read logs: %w", err)
			}

			return fmt.Errorf("check s3 storage: upload failed, %s", strings.Join(deconsolify(logsBytes), "\n"))
		}
	}

	return nil
}
