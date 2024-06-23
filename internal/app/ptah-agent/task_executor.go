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
	case *t.CreateNetworkReq:
		return e.createDockerNetwork(ctx, task.(*t.CreateNetworkReq))
	case *t.InitSwarmReq:
		return e.initSwarm(ctx, task.(*t.InitSwarmReq))
	case *t.CreateConfigReq:
		return e.createDockerConfig(ctx, task.(*t.CreateConfigReq))
	case *t.CreateSecretReq:
		return e.createDockerSecret(ctx, task.(*t.CreateSecretReq))
	case *t.CreateServiceReq:
		return e.createDockerService(ctx, task.(*t.CreateServiceReq))
	default:
		return nil, fmt.Errorf("execute task: unknown task type %T", task)
	}
}
