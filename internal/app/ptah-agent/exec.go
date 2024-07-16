package ptah_agent

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
	"io"
)

func (e *taskExecutor) exec(ctx context.Context, req *t.ServiceExecReq) (*t.ServiceExecRes, error) {
	containers, err := e.docker.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("com.docker.swarm.service.name=%s", req.ProcessName))),
	})

	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("no running containers found for service %s", req.ProcessName)
	}

	containerID := containers[0].ID

	idResponse, err := e.docker.ContainerExecCreate(ctx, containerID, req.ExecSpec)
	if err != nil {
		return nil, fmt.Errorf("create exec: %w", err)
	}

	response, err := e.docker.ContainerExecAttach(ctx, idResponse.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("attach exec: %w", err)
	}
	defer response.Close()

	output, err := io.ReadAll(response.Reader)
	if err != nil {
		return nil, fmt.Errorf("copy exec: %w", err)
	}

	res := t.ServiceExecRes{
		Output: deconsolify(output),
	}

	return &res, nil
}
