package ptah_agent

import (
	"context"
	dockerClient "github.com/docker/docker/client"
	"github.com/ptah-sh/ptah-agent/internal/pkg/networks"
	ptahClient "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
	"log"
	"time"
)

type Agent struct {
	version string
	ptah    *ptahClient.Client
	docker  *dockerClient.Client
}

func New(version string, baseUrl string, ptahToken string) *Agent {
	return &Agent{
		version: version,
		ptah:    ptahClient.New(baseUrl, ptahToken),
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

	startedReq := ptahClient.StartedReq{
		Version: a.version,
	}

	startedReq.Docker.Platform.Name = info.Name
	startedReq.Host.Networks = nets

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
		docker: a.docker,
	}

	log.Println("connected to server, poll interval", settings.Settings.PollInterval)

	for {
		time.Sleep(time.Duration(settings.Settings.PollInterval) * time.Second)

		nextTaskRes, err := a.ptah.GetNextTask(ctx)
		if err != nil {
			log.Println("can't get the next task", err)

			continue
		}

		if nextTaskRes == nil {
			continue
		}

		task, err := parseTask(nextTaskRes.TaskType, nextTaskRes.Payload)
		if err != nil {
			log.Println("can't parse task", err)

			continue
		}

		result, err := executor.executeTask(ctx, task)
		// TODO: store the result to re-send it once connection to the ptah server is restored
		if err == nil {
			if err = a.ptah.CompleteTask(ctx, nextTaskRes.ID, result); err != nil {
				log.Println("can't complete task", err)
			}
		} else {
			if err = a.ptah.FailTask(ctx, nextTaskRes.ID, &ptahClient.TaskError{
				Message: err.Error(),
			}); err != nil {
				log.Println("can't fail task", err)
			}
		}
	}
}
