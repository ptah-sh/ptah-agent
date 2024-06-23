package ptah_agent

import (
	"context"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) createDockerConfig(ctx context.Context, req *t.CreateConfigReq) (*t.CreateConfigRes, error) {
	var res t.CreateConfigRes

	response, err := e.docker.ConfigCreate(ctx, req.Payload.SwarmConfigSpec)
	if err != nil {
		return nil, err
	}

	res.Docker.ID = response.ID

	return &res, nil
}
