package ptah_agent

import (
	"context"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"

	"github.com/docker/docker/api/types/swarm"
)

func (e *taskExecutor) initSwarm(ctx context.Context, req t.InitSwarmReq) (*t.InitSwarmRes, error) {
	var res t.InitSwarmRes

	swarmId, err := e.docker.SwarmInit(ctx, swarm.InitRequest{
		ListenAddr:      "0.0.0.0:2377",
		AdvertiseAddr:   req.Payload.AdvertiseAddr,
		ForceNewCluster: req.Payload.Force,
	})

	if err != nil {
		return nil, err
	}

	res.Docker.ID = swarmId

	return &res, nil
}
