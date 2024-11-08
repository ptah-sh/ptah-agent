package registry

import "net/http"

type Registry struct {
	baseUrl string
	client  *http.Client
}

func New(baseUrl string) *Registry {
	return &Registry{
		baseUrl: baseUrl,
		client:  &http.Client{},
	}
}
