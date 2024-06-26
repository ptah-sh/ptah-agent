package ptah_agent

import (
	"context"
	"github.com/pkg/errors"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) updateCurrentNode(ctx context.Context, req *t.UpdateCurrentNodeReq) (*t.UpdateCurrentNodeRes, error) {
	var res t.UpdateCurrentNodeRes

	info, err := e.docker.Info(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "updateCurrentNode: docker.info() failed")
	}

	node, _, err := e.docker.NodeInspectWithRaw(ctx, info.Swarm.NodeID)
	if err != nil {
		return nil, errors.Wrapf(err, "updateCurrentNode: docker.NodeInspectWithRaw() failed")
	}

	err = e.docker.NodeUpdate(ctx, info.Swarm.NodeID, node.Version, req.NodeSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "updateCurrentNode: docker.NodeUpdate() failed")
	}

	return &res, nil
}
