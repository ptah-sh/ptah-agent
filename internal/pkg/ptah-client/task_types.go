package ptah_client

import (
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
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
	AuthConfigName string
	ReleaseCommand struct {
		ConfigName   string
		ConfigLabels map[string]string
		Command      string
	}
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

type DownloadAgentUpgradeReq struct {
	TargetVersion string
	DownloadUrl   string
}

type DownloadAgentUpgradeRes struct {
	FileSize     int `json:"fileSize"`
	DownloadTime int `json:"downloadTime"`
}

type UpdateAgentSymlinkReq struct {
	TargetVersion string
}

type UpdateAgentSymlinkRes struct {
}

type ConfirmAgentUpgradeReq struct {
	TargetVersion string
}

type ConfirmAgentUpgradeRes struct {
}

type CreateRegistryAuthReq struct {
	PrevConfigName  string
	AuthConfigSpec  registry.AuthConfig
	SwarmConfigSpec swarm.ConfigSpec
}

type CreateRegistryAuthRes struct {
	dockerIdRes
}

type CheckRegistryAuthReq struct {
	RegistryConfigName string
}

type CheckRegistryAuthRes struct {
	Status string
}

type PullImageReq struct {
	AuthConfigName  string
	Image           string
	PullOptionsSpec image.PullOptions
}

type PullImageRes struct {
	Output []string `json:"output"`
}
