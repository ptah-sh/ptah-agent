package ptah_agent

import (
	"context"
	"fmt"

	dockerClient "github.com/docker/docker/client"
	caddyClient "github.com/ptah-sh/ptah-agent/internal/pkg/caddy-client"

	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

type taskExecutor struct {
	safeClient *SafeClient
	agent      *Agent
	docker     *dockerClient.Client
	caddy      *caddyClient.Client
	rootDir    string
}

func (e *taskExecutor) executeTask(ctx context.Context, anyTask interface{}) (interface{}, error) {
	switch task := anyTask.(type) {
	case *t.CreateNetworkReq:
		return e.createDockerNetwork(ctx, task)
	case *t.InitSwarmReq:
		return e.initSwarm(ctx, task)
	case *t.CreateConfigReq:
		return e.createDockerConfig(ctx, task)
	case *t.CreateSecretReq:
		return e.createDockerSecret(ctx, task)
	case *t.CreateServiceReq:
		return e.createDockerService(ctx, task)
	case *t.ApplyCaddyConfigReq:
		return e.applyCaddyConfig(ctx, task)
	case *t.UpdateServiceReq:
		return e.updateDockerService(ctx, task)
	case *t.UpdateCurrentNodeReq:
		return e.updateCurrentNode(ctx, task)
	case *t.DeleteServiceReq:
		return e.deleteDockerService(ctx, task)
	case *t.DownloadAgentUpgradeReq:
		return e.downloadAgentUpgrade(ctx, task)
	case *t.UpdateAgentSymlinkReq:
		return e.updateAgentSymlink(ctx, task)
	case *t.ConfirmAgentUpgradeReq:
		return e.confirmAgentUpgrade(ctx, task)
	case *t.CreateRegistryAuthReq:
		return e.createRegistryAuth(ctx, task)
	case *t.CheckRegistryAuthReq:
		return e.checkRegistryAuth(ctx, task)
	case *t.PullImageReq:
		return e.pullImage(ctx, task)
	case *t.CreateS3StorageReq:
		return e.createS3Storage(ctx, task)
	case *t.CheckS3StorageReq:
		return e.checkS3Storage(ctx, task)
	case *t.ServiceExecReq:
		return e.exec(ctx, task)
	case *t.S3UploadReq:
		return e.s3upload(ctx, task)
	case *t.JoinSwarmReq:
		return e.joinSwarm(ctx, task)
	case *t.UpdateDirdReq:
		return e.updateDird(ctx, task)
	case *t.LaunchServiceReq:
		return e.launchDockerService(ctx, task)
	case *t.S3DownloadReq:
		return e.s3download(ctx, task)
	case *t.S3RemoveReq:
		return e.removeS3Files(ctx, task)
	case *t.PullGitRepoReq:
		return e.pullGitRepo(ctx, task)
	case *t.BuildImageReq:
		return e.buildImage(ctx, task)
	default:
		return nil, fmt.Errorf("execute task: unknown task type %T", task)
	}
}
