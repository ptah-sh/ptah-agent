package container

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
)

// ReadLogs reads and parses container logs from Docker daemon
func ReadLogs(ctx context.Context, docker *dockerClient.Client, containerID string) ([]string, error) {
	logs, err := docker.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	// Docker log format includes 8 bytes of header per line
	// First byte is stream type (1 = stdout, 2 = stderr)
	// Next 3 bytes are reserved (padding)
	// Last 4 bytes are length (uint32)
	var lines []string
	header := make([]byte, 8)
	for {
		_, err := io.ReadFull(logs, header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read log header: %w", err)
		}

		// Get the length of the log line from the header
		length := binary.BigEndian.Uint32(header[4:])

		// Read the log line
		line := make([]byte, length)
		_, err = io.ReadFull(logs, line)
		if err != nil {
			return nil, fmt.Errorf("failed to read log line: %w", err)
		}

		lines = append(lines, strings.TrimSpace(string(line)))
	}

	return lines, nil
}

// ReadLogsAsString reads container logs and returns them as a single string
func ReadLogsAsString(ctx context.Context, docker *dockerClient.Client, containerID string) (string, error) {
	lines, err := ReadLogs(ctx, docker, containerID)
	if err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}
