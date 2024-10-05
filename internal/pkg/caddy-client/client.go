package caddy_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type Client struct {
	url  string
	http *http.Client
}

func New(url string, http *http.Client) *Client {
	return &Client{
		url:  url,
		http: http,
	}
}

func (c *Client) GetMetrics(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/metrics", c.url), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "caddy: failed to create metrics request")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "caddy: failed to get metrics")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("caddy: failed to get metrics, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "caddy: failed to read metrics response body")
	}

	return strings.Split(string(body), "\n"), nil
}

func (c *Client) PostConfig(ctx context.Context, config map[string]interface{}) error {
	if config == nil {
		return fmt.Errorf("caddy: config is nil")
	}

	body, err := json.Marshal(config)
	if err != nil {
		return errors.Wrapf(err, "caddy: failed to marshal caddy config")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/config/", c.url), bytes.NewBuffer(body))
	if err != nil {
		return errors.Wrapf(err, "caddy: failed to POST %s/config/", c.url)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return errors.Wrapf(err, "caddy: failed to execute request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrapf(err, "caddy: failed to read response body")
		}

		var errMessage struct {
			Error string `json:"error"`
		}

		err = json.Unmarshal(respBody, &errMessage)
		if err != nil {
			return errors.Wrapf(err, "caddy: failed to unmarshal response body")
		}

		return fmt.Errorf("caddy error: %s", errMessage.Error)
	}

	return nil
}
