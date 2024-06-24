package ptah_client

import (
	"github.com/docker/docker/api/types/swarm"
)
import "github.com/docker/docker/api/types/network"

type TaskError struct {
	Message string `json:"message"`
}

type taskReq struct {
	ID int `json:"id"`
}

type dockerIdRes struct {
	Docker struct {
		ID string `json:"id"`
	} `json:"docker"`
}

type CreateNetworkReq struct {
	taskReq

	Payload struct {
		NetworkName          string
		NetworkCreateOptions network.CreateOptions
	}
}

type CreateNetworkRes struct {
	dockerIdRes
}

type InitSwarmReq struct {
	taskReq

	Payload struct {
		SwarmInitRequest swarm.InitRequest
	}
}

type InitSwarmRes struct {
	dockerIdRes
}

type CreateConfigReq struct {
	taskReq

	Payload struct {
		SwarmConfigSpec swarm.ConfigSpec
	}
}

type CreateConfigRes struct {
	dockerIdRes
}

type CreateSecretReq struct {
	taskReq

	Payload struct {
		SwarmSecretSpec swarm.SecretSpec
	}
}

type CreateSecretRes struct {
	dockerIdRes
}

type CreateServiceReq struct {
	taskReq

	Payload struct {
		AuthConfigName       string
		SecretVarsConfigName string
		SwarmServiceSpec     swarm.ServiceSpec
	}
}

type CreateServiceRes struct {
	dockerIdRes
}

type ApplyCaddyConfigReq struct {
	taskReq

	Payload struct {
		Caddy map[string]interface{} `json:"caddy"`
	} `json:"payload"`
}

type ApplyCaddyConfigRes struct {
}
