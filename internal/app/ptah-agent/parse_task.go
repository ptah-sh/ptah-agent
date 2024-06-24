package ptah_agent

import (
	"encoding/json"
	"fmt"
	ptahClient "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

// God Save Us.
func parseTask(taskType int, payload []byte) (interface{}, error) {
	switch taskType {
	case 0:
		var req ptahClient.CreateNetworkReq

		return unmarshalTask(payload, &req)
	case 1:
		var req ptahClient.InitSwarmReq

		return unmarshalTask(payload, &req)
	case 2:
		var req ptahClient.CreateConfigReq

		return unmarshalTask(payload, &req)
	case 3:
		var req ptahClient.CreateSecretReq

		return unmarshalTask(payload, &req)
	case 4:
		var req ptahClient.CreateServiceReq

		return unmarshalTask(payload, &req)
	case 5:
		var req ptahClient.ApplyCaddyConfigReq

		return unmarshalTask(payload, &req)
	default:
		return nil, fmt.Errorf("parse task: unknown task type %d", taskType)
	}
}

func unmarshalTask(payload []byte, task interface{}) (interface{}, error) {
	err := json.Unmarshal(payload, &task)
	if err != nil {
		return nil, err
	}

	return task, nil
}
