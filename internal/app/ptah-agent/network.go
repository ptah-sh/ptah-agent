package ptah_agent

import (
	"context"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) createDockerNetwork(ctx context.Context, req *t.CreateNetworkReq) (*t.CreateNetworkRes, error) {
	var res t.CreateNetworkRes

	response, err := e.docker.NetworkCreate(ctx, req.NetworkName, req.NetworkCreateOptions)
	if err != nil {
		return nil, err
	}

	res.Docker.ID = response.ID

	return &res, nil
}
