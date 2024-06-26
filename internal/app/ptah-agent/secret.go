package ptah_agent

import (
	"context"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) createDockerSecret(ctx context.Context, req *t.CreateSecretReq) (*t.CreateSecretRes, error) {
	var res t.CreateSecretRes

	response, err := e.docker.SecretCreate(ctx, req.SwarmSecretSpec)
	if err != nil {
		return nil, err
	}

	res.Docker.ID = response.ID

	return &res, nil
}
