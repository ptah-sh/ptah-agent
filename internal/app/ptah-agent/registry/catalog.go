package registry

import (
	"context"
	"encoding/json"
	"fmt"
)

type CatalogRes struct {
	Repositories []string `json:"repositories"`
}

func (r *Registry) Catalog(ctx context.Context) (*CatalogRes, error) {
	url := fmt.Sprintf("%s/v2/_catalog", r.baseUrl)

	resp, err := r.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get catalog: %w", err)
	}

	defer resp.Body.Close()

	var result CatalogRes
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode catalog: %w", err)
	}

	return &result, nil
}
