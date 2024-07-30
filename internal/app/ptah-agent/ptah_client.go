package ptah_agent

import (
	"context"
	"log"
	"net/http"
	"time"

	dockerClient "github.com/docker/docker/client"
	"github.com/pkg/errors"
	caddyClient "github.com/ptah-sh/ptah-agent/internal/pkg/caddy-client"
	"github.com/ptah-sh/ptah-agent/internal/pkg/networks"
	ptahClient "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

type Agent struct {
	Version string
	ptah    *ptahClient.Client
	rootDir string
	docker  *dockerClient.Client
	caddy   *caddyClient.Client
}

func New(version string, baseUrl string, ptahToken string, rootDir string) *Agent {
	return &Agent{
		Version: version,
		ptah:    ptahClient.New(baseUrl, ptahToken),
		rootDir: rootDir,
		caddy:   caddyClient.New("http://127.0.0.1:2019", http.DefaultClient),
	}
}

func (a *Agent) sendStartedEvent(ctx context.Context) (*ptahClient.StartedRes, error) {
	nets, err := networks.List()
	if err != nil {
		return nil, nil
	}

	docker, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		return nil, err
	}

	a.docker = docker

	a.docker.NegotiateAPIVersion(ctx)

	info, err := a.docker.Info(ctx)
	if err != nil {
		return nil, err
	}

	nodeData := ptahClient.NodeData{
		Version: a.Version,
	}

	nodeData.Docker.Platform.Name = info.Name
	nodeData.Host.Networks = nets
	if info.Swarm.ControlAvailable {
		nodeData.Role = "manager"
	} else {
		nodeData.Role = "worker"
	}

	startedReq := ptahClient.StartedReq{
		NodeData:  nodeData,
		SwarmData: nil,
	}

	if info.Swarm.NodeID != "" {
		swarm, err := a.docker.SwarmInspect(ctx)
		if err != nil {
			return nil, err
		}

		managerNodes := make([]ptahClient.ManagerNode, 0, len(info.Swarm.RemoteManagers))
		for _, manager := range info.Swarm.RemoteManagers {
			managerNodes = append(managerNodes, ptahClient.ManagerNode{
				NodeID: manager.NodeID,
				Addr:   manager.Addr,
			})
		}

		startedReq.SwarmData = &ptahClient.SwarmData{
			JoinTokens: ptahClient.JoinTokens{
				Worker:  swarm.JoinTokens.Worker,
				Manager: swarm.JoinTokens.Manager,
			},
			ManagerNodes: managerNodes,
		}
	}

	log.Println("sending started event, base url", a.ptah.BaseUrl)
	settings, err := a.ptah.Started(ctx, startedReq)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (a *Agent) Start(ctx context.Context) error {
	settings, err := a.sendStartedEvent(ctx)
	if err != nil {
		return err
	}

	executor := &taskExecutor{
		docker:  a.docker,
		caddy:   a.caddy,
		rootDir: a.rootDir,
		// TODO: use channel instead?
		stopAgentFlag: false,
		agent:         a,
	}

	log.Println("connected to server, poll interval", settings.Settings.PollInterval)

	for {
		taskID, task, err := a.getNextTask(ctx)
		if err != nil {
			log.Println("can't get the next task", err)
		}

		if task == nil {
			time.Sleep(time.Duration(settings.Settings.PollInterval) * time.Second)

			continue
		}

		result, err := executor.executeTask(ctx, task)
		// TODO: store the result to re-send it once connection to the ptah server is restored
		if err == nil {
			if err = a.ptah.CompleteTask(ctx, taskID, result); err != nil {
				log.Println("can't complete task", err)
			}
		} else {
			if err = a.ptah.FailTask(ctx, taskID, &ptahClient.TaskError{
				Message: err.Error(),
			}); err != nil {
				log.Println("can't fail task", err)
			}
		}

		if executor.stopAgentFlag {
			log.Println("received stop signal, shutting down gracefully")

			break
		}
	}

	return nil
}

func (a *Agent) getNextTask(ctx context.Context) (taskId int, task interface{}, err error) {
	nextTaskRes, err := a.ptah.GetNextTask(ctx)
	if err != nil {
		return 0, nil, errors.Wrapf(err, "agent.getNextTask")
	}

	if nextTaskRes == nil {
		return 0, nil, nil
	}

	task, err = parseTask(nextTaskRes.TaskType, nextTaskRes.Payload)
	if err != nil {
		return 0, nil, errors.Wrapf(err, "agent.getNextTask: parse task %d failed", nextTaskRes.TaskType)
	}

	return nextTaskRes.ID, task, nil
}
