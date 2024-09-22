package ptah_agent

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/pkg/errors"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) createDockerSecret(ctx context.Context, req *t.CreateSecretReq) (*t.CreateSecretRes, error) {
	var res t.CreateSecretRes

	decryptedData, err := e.decryptValue(ctx, string(req.SwarmSecretSpec.Data))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt secret data")
	}

	decryptedSpec := req.SwarmSecretSpec
	decryptedSpec.Data = []byte(decryptedData)

	response, err := e.docker.SecretCreate(ctx, decryptedSpec)
	if err != nil {
		return nil, err
	}

	res.Docker.ID = response.ID

	return &res, nil
}

func (e *taskExecutor) getSecretByName(ctx context.Context, name string) (*swarm.Secret, error) {
	if name == "" {
		return nil, errors.Wrapf(ErrSecretNotFound, "secret name is empty")
	}

	secrets, err := e.docker.SecretList(ctx, types.SecretListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", name),
		),
	})

	if err != nil {
		return nil, err
	}

	for _, secret := range secrets {
		if secret.Spec.Name == name {
			return &secret, nil
		}
	}

	return nil, errors.Wrapf(ErrSecretNotFound, "secret with name %s not found", name)
}
