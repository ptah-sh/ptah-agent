package ptah_agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

//go:embed s3.sh
var s3Script string

func (e *taskExecutor) createS3Storage(ctx context.Context, req *t.CreateS3StorageReq) (*t.CreateS3StorageRes, error) {
	var res t.CreateS3StorageRes

	decryptedSecretKey, err := e.decryptValue(ctx, req.S3StorageSpec.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("create s3 storage: decrypt secret key: %w", err)
	}

	decryptedSpec := req.S3StorageSpec
	decryptedSpec.SecretKey = decryptedSecretKey

	data, err := json.Marshal(decryptedSpec)
	if err != nil {
		return nil, fmt.Errorf("create s3 storage: marshal spec: %w", err)
	}

	req.SwarmConfigSpec.Data = data

	config, err := e.docker.ConfigCreate(ctx, req.SwarmConfigSpec)
	if err != nil {
		return nil, fmt.Errorf("create s3 storage: create config: %w", err)
	}

	res.Docker.ID = config.ID

	return &res, nil
}

func (e *taskExecutor) checkS3Storage(ctx context.Context, req *t.CheckS3StorageReq) (*t.CheckS3StorageRes, error) {
	_, err := e.uploadS3FileWithHelper(ctx, []mount.Mount{}, t.ArchiveSpec{}, false, req.S3StorageConfigName, "/tmp/check-access.txt", ".check-access")
	if err != nil {
		return nil, err
	}

	return &t.CheckS3StorageRes{}, nil
}

func (e *taskExecutor) s3upload(ctx context.Context, req *t.S3UploadReq) (*t.S3UploadRes, error) {
	output, err := e.uploadS3FileWithHelper(ctx, []mount.Mount{req.VolumeSpec}, req.Archive, req.RemoveSrcFile, req.S3StorageConfigName, req.SrcFilePath, req.DestFilePath)
	if err != nil {
		return nil, err
	}

	return &t.S3UploadRes{
		Output: output,
	}, nil
}

func (e *taskExecutor) uploadS3FileWithHelper(ctx context.Context, mounts []mount.Mount, archiveSpec t.ArchiveSpec, removeSrcFile bool, s3StorageConfigName, srcFilePath, destFilePath string) ([]string, error) {
	credentialsConfig, err := e.getConfigByName(ctx, s3StorageConfigName)
	if err != nil {
		return nil, fmt.Errorf("check s3 storage: get config: %w", err)
	}

	var s3StorageSpec t.S3StorageSpec
	err = json.Unmarshal(credentialsConfig.Spec.Data, &s3StorageSpec)
	if err != nil {
		return nil, fmt.Errorf("check s3 storage: unmarshal config: %w", err)
	}

	containerName := fmt.Sprintf("ptah-helper-%d", time.Now().Unix())

	if strings.Index(destFilePath, s3StorageSpec.PathPrefix) == 0 {
		destFilePath = destFilePath[len(s3StorageSpec.PathPrefix):]
	}

	// Pull the d3fk/s3cmd image before starting the container
	pull, err := e.docker.ImagePull(ctx, "d3fk/s3cmd", image.PullOptions{})
	if err != nil {
		return nil, fmt.Errorf("check s3 storage: pull image: %w", err)
	}
	defer pull.Close()

	_, err = io.ReadAll(pull)
	if err != nil {
		return nil, fmt.Errorf("check s3 storage: read image pull: %w", err)
	}

	createResponse, err := e.docker.ContainerCreate(ctx, &container.Config{
		User:  "root",
		Image: "d3fk/s3cmd",
		Entrypoint: []string{
			"sh",
			"-c",
		},
		Cmd: []string{
			s3Script,
		},
		Env: []string{
			fmt.Sprintf("ARCHIVE_ENABLED=%t", archiveSpec.Enabled),
			fmt.Sprintf("ARCHIVE_FORMAT=%s", archiveSpec.Format),
			fmt.Sprintf("S3_ACCESS_KEY=%s", s3StorageSpec.AccessKey),
			fmt.Sprintf("S3_SECRET_KEY=%s", s3StorageSpec.SecretKey),
			fmt.Sprintf("S3_ENDPOINT=%s", s3StorageSpec.Endpoint),
			fmt.Sprintf("S3_REGION=%s", s3StorageSpec.Region),
			fmt.Sprintf("S3_BUCKET=%s", s3StorageSpec.Bucket),
			fmt.Sprintf("PATH_PREFIX=%s", strings.Trim(s3StorageSpec.PathPrefix, "/")),
			fmt.Sprintf("SRC_FILE_PATH=%s", srcFilePath),
			fmt.Sprintf("DEST_FILE_PATH=%s", strings.TrimPrefix(destFilePath, "/")),
			fmt.Sprintf("REMOVE_SRC_FILE=%t", removeSrcFile),
		},
	}, &container.HostConfig{
		Mounts: mounts,
	}, &network.NetworkingConfig{}, &v1.Platform{}, containerName)
	if err != nil {
		return nil, fmt.Errorf("check s3 storage: create container: %w", err)
	}

	defer func(docker *dockerClient.Client, ctx context.Context, containerID string, options container.RemoveOptions) {
		err = docker.ContainerRemove(ctx, containerID, options)
		if err != nil {
			log.Printf("failed to remove container %s: %v", containerID, err)
		}
	}(e.docker, ctx, createResponse.ID, container.RemoveOptions{
		Force: true,
	})

	err = e.docker.ContainerStart(ctx, createResponse.ID, container.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("check s3 storage: start container: %w", err)
	}

	// FIXME: make as a standalone function - copy-pasted into service_monitor.go
	waitChan, errChan := e.docker.ContainerWait(ctx, createResponse.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errChan:
		if err != nil {
			return nil, fmt.Errorf("check s3 storage: wait container: %w", err)
		}
	case w := <-waitChan:
		if w.Error != nil {
			return nil, fmt.Errorf("check s3 storage: wait container: %s", w.Error.Message)
		}

		logs, err := e.readContainerLogs(ctx, createResponse.ID)
		if err != nil {
			return nil, fmt.Errorf("check s3 storage failed with error %s, read logs failed too: %w", w.Error.Message, err)
		}

		// FIXME: transfer all (stdout + stderr, success and error) logs to the ptah-server once logging support is added
		if w.StatusCode != 0 {
			return nil, fmt.Errorf("check s3 storage: upload failed, %s", logs)
		}

		return strings.Split(logs, "\n"), nil
	}

	return nil, nil
}
