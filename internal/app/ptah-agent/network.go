package ptah_agent

import (
	"context"
	"github.com/docker/docker/api/types/network"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) createDockerNetwork(ctx context.Context, req t.CreateNetworkReq) (*t.CreateNetworkRes, error) {
	var res t.CreateNetworkRes

	//net, err := e.docker.NetworkInspect(ctx, req.Payload.Name, network.InspectOptions{})
	//if d.IsErrNotFound(err) {
	response, err := e.docker.NetworkCreate(ctx, req.Payload.Name, network.CreateOptions{})
	if err != nil {
		return nil, err
	}

	res.Docker.ID = response.ID

	//return &result, nil
	//} else if err != nil {
	//	return nil, err
	//}

	return &res, nil
}
