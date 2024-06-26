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
	NetworkName          string
	NetworkCreateOptions network.CreateOptions
}

type CreateNetworkRes struct {
	dockerIdRes
}

type InitSwarmReq struct {
	SwarmInitRequest swarm.InitRequest
}

type InitSwarmRes struct {
	dockerIdRes
}

type CreateConfigReq struct {
	SwarmConfigSpec swarm.ConfigSpec
}

type CreateConfigRes struct {
	dockerIdRes
}

type CreateSecretReq struct {
	SwarmSecretSpec swarm.SecretSpec
}

type CreateSecretRes struct {
	dockerIdRes
}

type ServicePayload struct {
	AuthConfigName   string
	SwarmServiceSpec swarm.ServiceSpec
}

type SecretVars struct {
	ConfigName   string
	ConfigLabels map[string]string
	Values       map[string]string

	Preserve           []string
	PreserveFromConfig string
}

type CreateServiceReq struct {
	ServicePayload

	SecretVars SecretVars
}

type CreateServiceRes struct {
	dockerIdRes
}

type UpdateServiceReq struct {
	ServicePayload

	SecretVars SecretVars
}

type UpdateServiceRes struct {
}

type ApplyCaddyConfigReq struct {
	Caddy map[string]interface{} `json:"caddy"`
}

type ApplyCaddyConfigRes struct {
}

type UpdateCurrentNodeReq struct {
	NodeSpec swarm.NodeSpec
}

type UpdateCurrentNodeRes struct {
}

type DeleteServiceReq struct {
	ServiceName string
}

type DeleteServiceRes struct {
}
