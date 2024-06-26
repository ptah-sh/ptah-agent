package ptah_agent

import (
	"context"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) applyCaddyConfig(ctx context.Context, req *t.ApplyCaddyConfigReq) (*t.ApplyCaddyConfigRes, error) {
	var res t.ApplyCaddyConfigRes

	err := e.caddy.PostConfig(ctx, req.Caddy)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
