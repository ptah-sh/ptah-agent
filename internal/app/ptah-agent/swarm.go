package ptah_agent

import (
	"context"
	"github.com/pkg/errors"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
	"log"
	"runtime"
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
