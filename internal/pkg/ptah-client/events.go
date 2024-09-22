package ptah_client

import (
	"github.com/ptah-sh/ptah-agent/internal/pkg/networks"
)

type JoinTokens struct {
	Worker  string `json:"worker"`
	Manager string `json:"manager"`
}

type ManagerNode struct {
	NodeID string `json:"nodeId"`
	Addr   string `json:"addr"`
}

type SwarmData struct {
	JoinTokens    JoinTokens    `json:"joinTokens"`
	ManagerNodes  []ManagerNode `json:"managerNodes"`
	EncryptionKey string        `json:"encryptionKey"`
}

type NodeData struct {
	Version string `json:"version"`
	Docker  struct {
		Platform struct {
			Name string `json:"name"`
		} `json:"platform"`
	} `json:"docker"`
	Host struct {
		Networks []networks.Network `json:"networks"`
	} `json:"host"`
	Role string `json:"role"`
	Addr string `json:"address"`
}

type StartedReq struct {
	NodeData  NodeData   `json:"node"`
	SwarmData *SwarmData `json:"swarm"`
}

type StartedRes struct {
	Settings struct {
		PollInterval int `json:"poll_interval"`
	} `json:"settings"`
}
