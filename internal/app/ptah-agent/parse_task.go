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
		var createNetwork ptahClient.CreateNetworkReq

		return unmarshalTask(payload, &createNetwork)
	case 1:
		var initSwarm ptahClient.InitSwarmReq

		return unmarshalTask(payload, &initSwarm)
	case 2:
		var createConfig ptahClient.CreateConfigReq

		return unmarshalTask(payload, &createConfig)
	case 3:
		var createSecret ptahClient.CreateSecretReq

		return unmarshalTask(payload, &createSecret)
	case 4:
		var createService ptahClient.CreateServiceReq

		return unmarshalTask(payload, &createService)
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
