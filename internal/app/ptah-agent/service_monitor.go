package ptah_agent

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/pkg/errors"
	ptahContainer "github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/docker/container"
)

func (e *taskExecutor) monitorServiceLaunch(ctx context.Context, service *swarm.Service) error {
	log := Logger(ctx)

	log.Debug("monitoring service launch", "labels", service.Spec.Labels)

	if service.Spec.Mode.ReplicatedJob != nil {
		if service.Spec.Mode.ReplicatedJob.MaxConcurrent == nil {
			return errors.New("max concurrent is not set")
		}

		if *service.Spec.Mode.ReplicatedJob.MaxConcurrent == 0 {
			return nil
		}

		return e.monitorJobServiceLaunch(ctx, service)
	}

	if service.Spec.Mode.Replicated != nil {
		if service.Spec.Mode.Replicated.Replicas == nil {
			return errors.New("replicas is not set")
		}

		if *service.Spec.Mode.Replicated.Replicas == 0 {
			return nil
		}

		return e.monitorDaemonServiceLaunch(ctx, service)
	}

	return errors.New("unknown service mode")
}

func (e *taskExecutor) monitorDaemonServiceLaunch(ctx context.Context, service *swarm.Service) error {
	log := Logger(ctx)

	log.Debug("monitoring daemon service launch")

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	// TODO: make timeout configurable
	timeout := time.After(time.Duration(5) * time.Minute)

	successfullChecks := 0

	// FIXME: check if the container is continiously restarting.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			log.Debug("timeout")

			return errors.New("timeout")
		case <-ticker.C:
			log.Debug("inspecting service")

			service, _, err := e.docker.ServiceInspectWithRaw(ctx, service.ID, types.ServiceInspectOptions{})
			if err != nil {
				return err
			}

			if service.UpdateStatus == nil {
				log.Debug("service inspected, update status is nil, checking tasks")

				tasks, err := e.docker.TaskList(ctx, types.TaskListOptions{
					Filters: filters.NewArgs(
						filters.Arg("service", service.ID),
					),
				})
				if err != nil {
					return err
				}

				if len(tasks) == 0 && service.Spec.Mode.Replicated.Replicas != nil && *service.Spec.Mode.Replicated.Replicas == 0 {
					return nil
				}

				tasksRunning := 0
				for _, t := range tasks {
					if t.Status.State == swarm.TaskStateRunning {
						tasksRunning++
					}
				}

				if tasksRunning == int(*service.Spec.Mode.Replicated.Replicas) {
					successfullChecks++
				} else {
					successfullChecks = 0
				}

				if successfullChecks >= 3 {
					log.Debug("service launched", "service_id", service.ID, "successfull_checks", successfullChecks)

					return nil
				}
			} else {
				log.Debug("service inspected", "state", service.UpdateStatus.State)

				switch service.UpdateStatus.State {
				case swarm.UpdateStatePaused:
					// FIXME: rollback the service?
					return errors.Errorf("service update paused: %s", service.UpdateStatus.Message)
				case swarm.UpdateStateCompleted:
					return nil
				case swarm.UpdateStateRollbackCompleted:
					return errors.Errorf("service update failed: %s", service.UpdateStatus.Message)
				}
			}
		}
	}
}

func (e *taskExecutor) monitorJobServiceLaunch(ctx context.Context, service *swarm.Service) error {
	log := Logger(ctx)

	log.Debug("monitoring job service launch")

	timeout := time.After(time.Duration(30) * time.Second)

	taskTicker := time.NewTicker(time.Second * 5)
	defer taskTicker.Stop()

	var containerID string
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			log.Debug("timeout")

			return errors.New("timeout")
		case <-taskTicker.C:
			tasks, err := e.docker.TaskList(ctx, types.TaskListOptions{
				Filters: filters.NewArgs(
					filters.Arg("service", service.ID),
					filters.Arg("label", "sh.ptah.cookie="+service.Spec.Labels["sh.ptah.cookie"]),
				),
			})
			if err != nil {
				return err
			}

			for _, t := range tasks {
				if t.JobIteration.Index == service.JobStatus.JobIteration.Index {
					if t.Status.Err != "" {
						logs, err := ptahContainer.ReadLogsAsString(ctx, e.docker, t.Status.ContainerStatus.ContainerID)
						if err != nil {
							return fmt.Errorf("task failed with error %s, read logs failed too: %w", t.Status.Err, err)
						}

						log.Debug("task failed", "error", t.Status.Err)

						return errors.Errorf("task %s failed: %s\n%s", t.ID, t.Status.Err, logs)
					}

					if t.Status.ContainerStatus != nil {
						containerID = t.Status.ContainerStatus.ContainerID
						break
					}
				}
			}
		}

		if containerID != "" {
			break
		}
	}

	if containerID == "" {
		return errors.Errorf("container with task not found")
	}

	log.Debug("container found", "containerID", containerID)

	waitChan, errChan := e.docker.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	// FIXME: make timeout configurable
	// timeout := time.After(time.Duration(10) * time.Minute)
	timeout = time.After(time.Duration(1) * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			log.Debug("timeout, stopping task")

			return errors.New("timeout")
		case err := <-errChan:
			log.Debug("error", "error", err)

			if err != nil {
				return fmt.Errorf("wait container: %w", err)
			}

			return nil
		case w := <-waitChan:
			if w.Error != nil {
				log.Debug("wait container", "error", w.Error.Message)

				return fmt.Errorf("wait container: %s", w.Error.Message)
			}

			// FIXME: transfer all (stdout + stderr, success and error) logs to the ptah-server once logging support is added
			if w.StatusCode != 0 {
				logs, err := ptahContainer.ReadLogsAsString(ctx, e.docker, containerID)
				if err != nil {
					return fmt.Errorf("task failed with error %s, read logs failed too: %w", w.Error.Message, err)
				}

				return fmt.Errorf("task failed, %s", logs)
			}

			return nil
		}
	}
}
