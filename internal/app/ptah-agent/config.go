package ptah_agent

import (
	"context"
	"fmt"

	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

var (
	ErrServiceNotFound = fmt.Errorf("service not found")
	ErrSecretNotFound  = fmt.Errorf("secret not found")
)

func (e *taskExecutor) createDockerConfig(ctx context.Context, req *t.CreateConfigReq) (*t.CreateConfigRes, error) {
	var res t.CreateConfigRes

	response, err := e.docker.ConfigCreate(ctx, req.SwarmConfigSpec)
	if err != nil {
		return nil, err
	}

	res.Docker.ID = response.ID

	return &res, nil
}
