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
		err := json.Unmarshal(payload, &createNetwork)
		if err != nil {
			return nil, err
		}

		return createNetwork, nil
	case 1:
		var initSwarm ptahClient.InitSwarmReq
		err := json.Unmarshal(payload, &initSwarm)
		if err != nil {
			return nil, err
		}

		return initSwarm, nil
	default:
		return nil, fmt.Errorf("unknown task type %d", taskType)
	}
}
