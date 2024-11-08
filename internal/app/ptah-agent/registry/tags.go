package registry

import (
	"context"
	"encoding/json"
	"fmt"
)

type TagsListRes struct {
	Tags []string `json:"tags"`
}

func (r *Registry) TagsList(ctx context.Context, repo string) (*TagsListRes, error) {
	url := fmt.Sprintf("%s/v2/%s/tags/list", r.baseUrl, repo)

	resp, err := r.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get tags list: %w", err)
	}

	defer resp.Body.Close()

	var result TagsListRes
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode tags list: %w", err)
	}

	return &result, nil
}
