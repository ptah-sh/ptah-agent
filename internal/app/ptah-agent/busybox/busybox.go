package busybox

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	ptahContainer "github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/docker/container"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/encryption"
)

const busyboxImage = "ghcr.io/ptah-sh/ptah-busybox:latest"

type Config struct {
	Cmd     string
	Mounts  []mount.Mount
	EnvVars []string
}

type busybox struct {
	docker      *dockerClient.Client
	containerID string
}

func New(docker *dockerClient.Client) *busybox {
	return &busybox{
		docker: docker,
	}
}

func (b *busybox) Start(ctx context.Context, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}

	if cfg.Cmd == "" {
		return fmt.Errorf("cmd is required")
	}

	// Check if container is already running
	if b.containerID != "" {
		return fmt.Errorf("busybox container is already running with ID: %s", b.containerID)
	}

	// Pull the latest busybox image
	pull, err := b.docker.ImagePull(ctx, busyboxImage, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull busybox image: %w", err)
	}

	_, err = io.ReadAll(pull)
	if err != nil {
		return fmt.Errorf("failed to pull busybox image: %w", err)
	}

	// Get encryption keys
	keyPair, err := encryption.GetSshKeyPair(ctx, b.docker)
	if err != nil {
		return fmt.Errorf("failed to get encryption keys: %w", err)
	}

	envVars := cfg.EnvVars
	envVars = append(envVars, fmt.Sprintf("PTAH_SSH_PUBLIC_KEY=%s", keyPair.PublicKey))
	envVars = append(envVars, fmt.Sprintf("PTAH_SSH_PRIVATE_KEY=%s", keyPair.PrivateKey))

	// Prepare container configuration
	containerConfig := &container.Config{
		Image: busyboxImage,
		Entrypoint: []string{
			"/ptah/bin/entrypoint.sh",
		},
		Cmd: []string{
			cfg.Cmd,
		},
		Env: envVars,
	}

	// Add Docker socket mount to existing mounts
	dockerSocketMount := mount.Mount{
		Type:   mount.TypeBind,
		Source: "/var/run/docker.sock",
		Target: "/var/run/docker.sock",
	}

	mounts := append(cfg.Mounts, dockerSocketMount)

	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}

	// Configure network
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"ptah_net": {},
		},
	}

	containerName := fmt.Sprintf("ptah-busybox-%d", time.Now().Unix())

	// Create and start the container
	resp, err := b.docker.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, nil, containerName)
	if err != nil {
		return fmt.Errorf("failed to create busybox container: %w", err)
	}

	b.containerID = resp.ID

	if err := b.docker.ContainerStart(ctx, b.containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start busybox container: %w", err)
	}

	return nil
}

func (b *busybox) Stop(ctx context.Context) error {
	if b.containerID == "" {
		return nil
	}

	timeout := int(30) // 30 seconds timeout for stopping the container
	if err := b.docker.ContainerStop(ctx, b.containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to stop busybox container: %w", err)
	}

	if err := b.docker.ContainerRemove(ctx, b.containerID, container.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove busybox container: %w", err)
	}

	b.containerID = ""
	return nil
}

func (b *busybox) Wait(ctx context.Context) error {
	if b.containerID == "" {
		return fmt.Errorf("no busybox container is running")
	}

	statusCh, errCh := b.docker.ContainerWait(ctx, b.containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return fmt.Errorf("error waiting for busybox container: %w", err)
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("busybox container exited with non-zero status code: %d", status.StatusCode)
		}

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *busybox) Logs(ctx context.Context) ([]string, error) {
	if b.containerID == "" {
		return nil, fmt.Errorf("no busybox container is running")
	}

	return ptahContainer.ReadLogs(ctx, b.docker, b.containerID)
}
