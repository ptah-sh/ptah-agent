package ptah_agent

import (
	"encoding/json"
	"fmt"

	ptahClient "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func parseTask(taskType int, payload string) (interface{}, error) {
	switch taskType {

	case 0:
		return unmarshalTask(payload, &ptahClient.CreateNetworkReq{})
	case 1:
		return unmarshalTask(payload, &ptahClient.InitSwarmReq{})
	case 2:
		return unmarshalTask(payload, &ptahClient.CreateConfigReq{})
	case 3:
		return unmarshalTask(payload, &ptahClient.CreateSecretReq{})
	case 4:
		return unmarshalTask(payload, &ptahClient.CreateServiceReq{})
	case 5:
		return unmarshalTask(payload, &ptahClient.ApplyCaddyConfigReq{})
	case 6:
		return unmarshalTask(payload, &ptahClient.UpdateServiceReq{})
	case 7:
		return unmarshalTask(payload, &ptahClient.UpdateCurrentNodeReq{})
	case 8:
		return unmarshalTask(payload, &ptahClient.DeleteServiceReq{})
	case 9:
		return unmarshalTask(payload, &ptahClient.DownloadAgentUpgradeReq{})
	case 10:
		return unmarshalTask(payload, &ptahClient.UpdateAgentSymlinkReq{})
	case 11:
		return unmarshalTask(payload, &ptahClient.ConfirmAgentUpgradeReq{})
	case 12:
		return unmarshalTask(payload, &ptahClient.CreateRegistryAuthReq{})
	case 13:
		return unmarshalTask(payload, &ptahClient.CheckRegistryAuthReq{})
	case 14:
		return unmarshalTask(payload, &ptahClient.PullImageReq{})
	case 15:
		return unmarshalTask(payload, &ptahClient.CreateS3StorageReq{})
	case 16:
		return unmarshalTask(payload, &ptahClient.CheckS3StorageReq{})
	case 17:
		return unmarshalTask(payload, &ptahClient.ServiceExecReq{})
	case 18:
		return unmarshalTask(payload, &ptahClient.S3UploadReq{})
	case 19:
		return unmarshalTask(payload, &ptahClient.JoinSwarmReq{})
	case 20:
		return unmarshalTask(payload, &ptahClient.UpdateDirdReq{})
	case 21:
		return unmarshalTask(payload, &ptahClient.LaunchServiceReq{})
	case 22:
		return unmarshalTask(payload, &ptahClient.S3DownloadReq{})
	case 23:
		return unmarshalTask(payload, &ptahClient.S3RemoveReq{})
	case 24:
		return unmarshalTask(payload, &ptahClient.PullGitRepoReq{})
	case 25:
		return unmarshalTask(payload, &ptahClient.BuildImageReq{})
	default:
		return nil, fmt.Errorf("parse task: unknown task type %d", taskType)
	}
}

func unmarshalTask(payload string, task interface{}) (interface{}, error) {
	err := json.Unmarshal([]byte(payload), &task)
	if err != nil {
		return nil, err
	}

	return task, nil
}
