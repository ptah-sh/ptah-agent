package ptah_agent

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types/registry"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) createRegistryAuth(ctx context.Context, req *t.CreateRegistryAuthReq) (*t.CreateRegistryAuthRes, error) {
	if req.PrevConfigName != "" {
		prev, _, err := e.docker.ConfigInspectWithRaw(ctx, req.PrevConfigName)
		if err != nil {
			return nil, err
		}

		var authConfig registry.AuthConfig
		err = json.Unmarshal(prev.Spec.Data, &authConfig)
		if err != nil {
			return nil, err
		}

		if req.AuthConfigSpec.Username == "" {
			req.AuthConfigSpec.Username = authConfig.Username
		}

		if req.AuthConfigSpec.Password == "" {
			req.AuthConfigSpec.Password = authConfig.Password
		}
	}

	data, err := json.Marshal(req.AuthConfigSpec)
	if err != nil {
		return nil, err
	}

	req.SwarmConfigSpec.Data = data

	_, err = e.docker.ConfigCreate(ctx, req.SwarmConfigSpec)
	if err != nil {
		return nil, err
	}

	return &t.CreateRegistryAuthRes{}, nil
}

func (e *taskExecutor) checkRegistryAuth(ctx context.Context, req *t.CheckRegistryAuthReq) (*t.CheckRegistryAuthRes, error) {
	config, _, err := e.docker.ConfigInspectWithRaw(ctx, req.RegistryConfigName)
	if err != nil {
		return nil, err
	}

	var authConfig registry.AuthConfig
	err = json.Unmarshal(config.Spec.Data, &authConfig)
	if err != nil {
		return nil, err
	}

	resp, err := e.docker.RegistryLogin(ctx, authConfig)
	if err != nil {
		return nil, err
	}

	return &t.CheckRegistryAuthRes{Status: resp.Status}, nil
}
