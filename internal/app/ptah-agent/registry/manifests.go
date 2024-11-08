package registry

import (
	"context"
	"fmt"
	"net/http"
)

type ManifestHeadRes struct {
	Digest string `json:"digest"`
}

func (r *Registry) ManifestHead(ctx context.Context, repo, tag string) (*ManifestHeadRes, error) {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", r.baseUrl, repo, tag)

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create manifest head request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do manifest head request: %w", err)
	}

	defer resp.Body.Close()

	return &ManifestHeadRes{
		Digest: resp.Header.Get("Docker-Content-Digest"),
	}, nil
}

func (r *Registry) ManifestDelete(ctx context.Context, repo, digest string) error {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", r.baseUrl, repo, digest)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create manifest delete request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("do manifest delete request: %w", err)
	}

	defer resp.Body.Close()

	return nil
}
