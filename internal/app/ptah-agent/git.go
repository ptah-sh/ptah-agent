package ptah_agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/busybox"
	ptah_client "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) pullGitRepo(ctx context.Context, task *ptah_client.PullGitRepoReq) (*ptah_client.PullGitRepoRes, error) {
	var result ptah_client.PullGitRepoRes

	box := busybox.New(e.docker)

	err := box.Start(ctx, &busybox.Config{
		EnvVars: []string{
			fmt.Sprintf("GIT_REPO=%s", task.Repo),
			fmt.Sprintf("GIT_REF=%s", task.Ref),
			fmt.Sprintf("TARGET_DIR=%s", task.TargetDir),
		},
		Mounts: []mount.Mount{task.VolumeSpec},
		Cmd:    "/ptah/bin/git_pull.sh",
	})
	if err != nil {
		return nil, errors.Wrap(err, "pull git repo")
	}

	defer box.Stop(ctx)

	err = box.Wait(ctx)
	if err != nil {
		logs, logsErr := box.Logs(ctx)
		if logsErr != nil {
			return nil, errors.Wrap(logsErr, "get git pull logs")
		}

		message := strings.Join(logs, "\n")

		return nil, errors.Wrapf(err, "%s\nwait for git pull", message)
	}

	logs, err := box.Logs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get git pull logs")
	}

	result.Output = logs

	return &result, nil
}
