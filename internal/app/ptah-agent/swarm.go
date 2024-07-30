package ptah_agent

import (
	"context"
	"log"
	"runtime"

	"github.com/pkg/errors"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) initSwarm(ctx context.Context, req *t.InitSwarmReq) (*t.InitSwarmRes, error) {
	var res t.InitSwarmRes

	if runtime.GOOS == "darwin" {
		log.Println("Docker Desktop on MacOS doesn't support AdvertiseAddr, using 127.0.0.1:2377")

		req.SwarmInitRequest.AdvertiseAddr = "127.0.0.1:2377"
	}

	swarmId, err := e.docker.SwarmInit(ctx, req.SwarmInitRequest)
	if err != nil {
		return nil, errors.Wrapf(err, "init swarm")
	}

	res.Docker.ID = swarmId

	return &res, nil
}

func (e *taskExecutor) joinSwarm(ctx context.Context, req *t.JoinSwarmReq) (*t.JoinSwarmRes, error) {
	var res t.JoinSwarmRes

	err := e.docker.SwarmJoin(ctx, req.JoinSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "join swarm")
	}

	return &res, nil
}
