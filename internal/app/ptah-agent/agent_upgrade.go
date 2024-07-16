package ptah_agent

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"time"
)

func (e *taskExecutor) downloadAgentUpgrade(ctx context.Context, req *t.DownloadAgentUpgradeReq) (*t.DownloadAgentUpgradeRes, error) {
	startTime := time.Now()

	request, err := http.NewRequestWithContext(ctx, "GET", req.DownloadUrl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/octet-stream")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download agent upgrade: unexpected status code %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/octet-stream" {
		return nil, fmt.Errorf("download agent upgrade: unexpected content type %s", resp.Header.Get("Content-Type"))
	}

	file, err := os.Create(path.Join(e.rootDir, "versions", req.TargetVersion))
	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	currentPerms := fileStat.Mode()
	// chmod +x
	newPerms := currentPerms | 0111

	err = file.Chmod(newPerms)
	if err != nil {
		return nil, err
	}

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		return nil, err
	}

	elapsedTime := time.Since(startTime).Seconds()

	return &t.DownloadAgentUpgradeRes{DownloadTime: int(math.Ceil(elapsedTime)), FileSize: int(written)}, nil
}

func (e *taskExecutor) updateAgentSymlink(ctx context.Context, req *t.UpdateAgentSymlinkReq) (*t.UpdateAgentSymlinkRes, error) {
	current := path.Join(e.rootDir, "current")

	err := os.Remove(current)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	err = os.Symlink(path.Join(e.rootDir, "versions", req.TargetVersion), current)
	if err != nil {
		return nil, err
	}

	e.stop()

	return &t.UpdateAgentSymlinkRes{}, nil
}

func (e *taskExecutor) confirmAgentUpgrade(ctx context.Context, req *t.ConfirmAgentUpgradeReq) (*t.ConfirmAgentUpgradeRes, error) {
	if e.agent.Version == req.TargetVersion {
		return &t.ConfirmAgentUpgradeRes{}, nil
	}

	return nil, fmt.Errorf("upgrade failed: current version %s, target version %s", e.agent.Version, req.TargetVersion)
}
