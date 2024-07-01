package ptah_agent

import (
	"context"
	"fmt"
	dockerClient "github.com/docker/docker/client"
	caddyClient "github.com/ptah-sh/ptah-agent/internal/pkg/caddy-client"
)
import (
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

type taskExecutor struct {
	agent         *Agent
	docker        *dockerClient.Client
	caddy         *caddyClient.Client
	rootDir       string
	stopAgentFlag bool
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
	case *t.ApplyCaddyConfigReq:
		return e.applyCaddyConfig(ctx, task.(*t.ApplyCaddyConfigReq))
	case *t.UpdateServiceReq:
		return e.updateDockerService(ctx, task.(*t.UpdateServiceReq))
	case *t.UpdateCurrentNodeReq:
		return e.updateCurrentNode(ctx, task.(*t.UpdateCurrentNodeReq))
	case *t.DeleteServiceReq:
		return e.deleteDockerService(ctx, task.(*t.DeleteServiceReq))
	case *t.DownloadAgentUpgradeReq:
		return e.downloadAgentUpgrade(ctx, task.(*t.DownloadAgentUpgradeReq))
	case *t.UpdateAgentSymlinkReq:
		return e.updateAgentSymlink(ctx, task.(*t.UpdateAgentSymlinkReq))
	case *t.ConfirmAgentUpgradeReq:
		return e.confirmAgentUpgrade(ctx, task.(*t.ConfirmAgentUpgradeReq))
	default:
		return nil, fmt.Errorf("execute task: unknown task type %T", task)
	}
}

func (e *taskExecutor) stop() {
	e.stopAgentFlag = true
}
