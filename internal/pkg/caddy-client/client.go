package caddy_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func (c *Client) PostConfig(ctx context.Context, config map[string]interface{}) error {
	if config == nil {
		return fmt.Errorf("caddy client: config is nil")
	}

	body, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/config/", c.url), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var errMessage struct {
			Error string `json:"error"`
		}

		err = json.Unmarshal(respBody, &errMessage)
		if err != nil {
			return err
		}

		return fmt.Errorf("caddy error: %s", errMessage.Error)
	}

	return nil
}
