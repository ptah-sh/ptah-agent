package ptah_client

import "github.com/ptah-sh/ptah-agent/internal/pkg/networks"

type StartedReq struct {
	Version string `json:"version"`
	Docker  struct {
		Platform struct {
			Name string `json:"name"`
		} `json:"platform"`
	} `json:"docker"`
	Host struct {
		Networks []networks.Network `json:"networks"`
	} `json:"host"`
}

type StartedRes struct {
	Settings struct {
		PollInterval int `json:"poll_interval"`
	} `json:"settings"`
}
