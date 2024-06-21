package ptah_client

type TaskError struct {
	Message string `json:"message"`
}

type task struct {
	ID int `json:"id"`
}

type dockerIdRes struct {
	Docker struct {
		ID string `json:"id"`
	} `json:"docker"`
}

type CreateNetworkReq struct {
	task

	Payload struct {
		Name string `json:"name"`
	}
}

type CreateNetworkRes struct {
	dockerIdRes
}

type InitSwarmReq struct {
	task

	Payload struct {
		Name          string `json:"name"`
		AdvertiseAddr string `json:"advertiseAddr"`
		Force         bool   `json:"forceNewCluster"`
	}
}

type InitSwarmRes struct {
	dockerIdRes
}
