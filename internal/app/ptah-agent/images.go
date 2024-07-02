package ptah_agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
	"io"
	"strings"
)

func (e *taskExecutor) pullImage(ctx context.Context, req *t.PullImageReq) (*t.PullImageRes, error) {
	if req.AuthConfigName != "" {
		config, err := e.getConfigByName(ctx, req.AuthConfigName)
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

	// Output is a json stream, but let's simplify things and start by collecting the whole outputStream at once.
	message, err := io.ReadAll(outputStream)
	if err != nil {
		return nil, fmt.Errorf("pull image: read outputStream: %w", err)
	}

	var line struct {
		Status string `json:"status"`
	}

	output := make([]string, 0)

	for _, row := range strings.Split(string(message), "\n") {
		row = strings.Trim(row, " \r\n")
		if row == "" {
			continue
		}

		if err := json.Unmarshal([]byte(row), &line); err != nil {
			return nil, fmt.Errorf("pull image: unmarshal json '%s': %w", row, err)
		}

		output = append(output, line.Status)
	}

	return &t.PullImageRes{Output: output}, nil
}
