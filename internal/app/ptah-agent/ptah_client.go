package ptah_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	dockerClient "github.com/docker/docker/client"
	"github.com/pkg/errors"
	caddyClient "github.com/ptah-sh/ptah-agent/internal/pkg/caddy-client"
	"github.com/ptah-sh/ptah-agent/internal/pkg/networks"
	ptahClient "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

type Agent struct {
	Version      string
	ptah         *ptahClient.Client
	safeClient   *SafeClient
	rootDir      string
	docker       *dockerClient.Client
	caddy        *caddyClient.Client
	executor     *taskExecutor
	metricsAgent *MetricsAgent
	cancel       context.CancelFunc
}

func New(version string, baseUrl string, ptahToken string, rootDir string) (*Agent, error) {
	docker, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("error initializing Docker client: %w", err)
	}

	// Create a background context for API version negotiation
	ctx := context.Background()

	docker.NegotiateAPIVersion(ctx)

	caddy := caddyClient.New("http://127.0.0.1:2019", http.DefaultClient)

	ptah := ptahClient.New(baseUrl, ptahToken)

	safeClient, err := NewSafeClient(ptah, rootDir)
	if err != nil {
		return nil, err
	}

	metricsAgent := NewMetricsAgent(safeClient, caddy, 5*time.Second)

	// TODO: refactor to avoid duplication and circular dependency?
	agent := &Agent{
		Version: version,
		// TODO: replace ptah with safeClient
		ptah:       ptah,
		safeClient: safeClient,
		rootDir:    rootDir,
		caddy:      caddy,
		docker:     docker,
		executor: &taskExecutor{
			docker:     docker,
			caddy:      caddy,
			rootDir:    rootDir,
			safeClient: safeClient,
		},
		metricsAgent: metricsAgent,
	}

	agent.executor.agent = agent

	return agent, nil
}

func (a *Agent) sendStartedEvent(ctx context.Context) (*ptahClient.StartedRes, error) {
	nets, err := networks.List()
	if err != nil {
		return nil, nil
	}

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

	if info.Swarm.NodeAddr != "" {
		nodeData.Addr = info.Swarm.NodeAddr
	} else {
		nodeData.Addr = nets[0].IPs[0].IP
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

		workerJoinToken, err := a.executor.encryptValue(ctx, swarm.JoinTokens.Worker)
		if err != nil {
			return nil, err
		}

		managerJoinToken, err := a.executor.encryptValue(ctx, swarm.JoinTokens.Manager)
		if err != nil {
			return nil, err
		}

		startedReq.SwarmData = &ptahClient.SwarmData{
			JoinTokens: ptahClient.JoinTokens{
				Worker:  workerJoinToken,
				Manager: managerJoinToken,
			},
			ManagerNodes: managerNodes,
		}

		encryptionKey, err := a.executor.getEncryptionKey(ctx)
		if err != nil {
			return nil, err
		}

		startedReq.SwarmData.EncryptionKey = encryptionKey.PublicKey
	}

	log.Println("sending started event, base url", a.ptah.BaseUrl)
	settings, err := a.ptah.Started(ctx, startedReq)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (a *Agent) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	a.cancel = cancel

	defer a.safeClient.Close()

	a.metricsAgent.Start(ctx)
	a.safeClient.StartBackgroundRequestsProcessing(ctx)

	settings, err := a.sendStartedEvent(ctx)
	if err != nil {
		return err
	}

	log.Println("connected to server, poll interval", settings.Settings.PollInterval)

	consecutiveFailures := 0
	maxConsecutiveFailures := 5

	for {
		select {
		case <-ctx.Done():
			log.Println("received stop signal, shutting down gracefully")

			return nil
		case <-time.After(time.Duration(settings.Settings.PollInterval) * time.Second):
			err = a.safeClient.PerformForegroundRequests(ctx)
			if err != nil {
				log.Println("can't perform calls", err)
				consecutiveFailures++
			}

			if consecutiveFailures >= maxConsecutiveFailures {
				return fmt.Errorf("shutting down due to %d consecutive failures performing calls", maxConsecutiveFailures)
			}

			if err != nil {
				continue
			}

			taskID, task, err := a.getNextTask(ctx)
			if err != nil {
				log.Println("can't get the next task", err)
				consecutiveFailures++

				if taskID == 0 {
					if consecutiveFailures >= maxConsecutiveFailures {
						return fmt.Errorf("shutting down due to %d consecutive failures to get next task", maxConsecutiveFailures)
					}
				} else {
					if err = a.safeClient.FailTask(ctx, taskID, &ptahClient.TaskError{
						Message: err.Error(),
					}); err != nil {
						log.Println("can't fail task", err)
					}
				}
			} else {
				consecutiveFailures = 0
			}

			if task == nil {
				continue
			}

			result, err := a.executor.executeTask(ctx, task)

			if err == nil {
				if err = a.safeClient.CompleteTask(ctx, taskID, result); err != nil {
					log.Println("can't complete task", err)
				}
			} else {
				if err = a.safeClient.FailTask(ctx, taskID, &ptahClient.TaskError{
					Message: err.Error(),
				}); err != nil {
					log.Println("can't fail task", err)
				}
			}
		}
	}
}

func (a *Agent) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
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
		return nextTaskRes.ID, nil, errors.Wrapf(err, "agent.getNextTask: parse task %d failed", nextTaskRes.TaskType)
	}

	return nextTaskRes.ID, task, nil
}

func (a *Agent) ExecTasks(ctx context.Context, jsonFilePath string) error {
	// Docker client should already be initialized and version negotiated in New()
	if a.docker == nil {
		return fmt.Errorf("docker client not initialized")
	}

	// Read the JSON file
	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	// Parse the JSON data
	var taskList []ptahClient.GetNextTaskRes
	err = json.Unmarshal(jsonData, &taskList)
	if err != nil {
		return fmt.Errorf("error parsing JSON data: %w", err)
	}

	executor := &taskExecutor{
		docker:  a.docker,
		caddy:   a.caddy,
		rootDir: a.rootDir,
		agent:   a,
	}

	// Execute each task
	for _, taskRes := range taskList {
		task, err := parseTask(taskRes.TaskType, taskRes.Payload)
		if err != nil {
			return fmt.Errorf("error parsing task %d: %w", taskRes.ID, err)
		}

		var result interface{}
		if taskRes.TaskType == 5 {
			result, err = a.retryTask(ctx, executor, task, 5*time.Second, 3*time.Minute)
		} else {
			result, err = executor.executeTask(ctx, task)
		}

		if err != nil {
			return fmt.Errorf("error executing task %d: %w", taskRes.ID, err)
		}

		log.Printf("Task %d (Type: %d) executed successfully. Result: %+v", taskRes.ID, taskRes.TaskType, result)
	}

	return nil
}

func (a *Agent) retryTask(ctx context.Context, executor *taskExecutor, task interface{}, retryInterval time.Duration, maxDuration time.Duration) (interface{}, error) {
	startTime := time.Now()
	for {
		result, err := executor.executeTask(ctx, task)
		if err == nil {
			return result, nil
		}

		if time.Since(startTime) >= maxDuration {
			return nil, fmt.Errorf("'Apply Caddy Config' task execution failed after %v: %w", maxDuration, err)
		}

		log.Printf("'Apply Caddy Config' task failed, retrying in %v: %v", retryInterval, err)
		time.Sleep(retryInterval)
	}
}
