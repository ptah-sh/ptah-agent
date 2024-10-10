package ptah_agent

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
)

func (e *taskExecutor) readContainerLogs(ctx context.Context, containerID string) (string, error) {
	logs, err := e.docker.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStderr: true,
		ShowStdout: true,
	})
	if err != nil {
		return "", fmt.Errorf("get logs: %w", err)
	}

	logsBytes, err := io.ReadAll(logs)
	if err != nil {
		return "", fmt.Errorf("read logs: %w", err)
	}

	return strings.Join(deconsolify(logsBytes), "\n"), nil
}
