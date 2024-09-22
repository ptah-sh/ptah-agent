package ptah_agent

import (
	"context"
	"encoding/json"

	"github.com/docker/docker/api/types/registry"
	"github.com/pkg/errors"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) createRegistryAuth(ctx context.Context, req *t.CreateRegistryAuthReq) (*t.CreateRegistryAuthRes, error) {
	decryptedPassword, err := e.decryptValue(ctx, req.AuthConfigSpec.Password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt password")
	}
	req.AuthConfigSpec.Password = decryptedPassword

	data, err := json.Marshal(req.AuthConfigSpec)
	if err != nil {
		return nil, err
	}

	req.SwarmConfigSpec.Data = data

	config, err := e.docker.ConfigCreate(ctx, req.SwarmConfigSpec)
	if err != nil {
		return nil, err
	}

	res := t.CreateRegistryAuthRes{}
	res.Docker.ID = config.ID

	return &res, nil
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
