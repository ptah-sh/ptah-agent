package ptah_agent

import (
	"context"
	"fmt"
	dockerClient "github.com/docker/docker/client"
)
import (
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

type taskExecutor struct {
	docker *dockerClient.Client
}

func (e *taskExecutor) executeTask(ctx context.Context, task interface{}) (interface{}, error) {
	switch task.(type) {
	case t.CreateNetworkReq:
		return e.createDockerNetwork(ctx, task.(t.CreateNetworkReq))
	case t.InitSwarmReq:
		return e.initSwarm(ctx, task.(t.InitSwarmReq))
	default:
		return nil, fmt.Errorf("unknown task type %T", task)
	}
}
