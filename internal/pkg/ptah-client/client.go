package ptah_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ServiceError struct {
	Message string `json:"message"`
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("ptah error: %s", e.Message)
}

type Client struct {
	BaseUrl   string
	ptahToken string
	http      *http.Client
}

func New(baseUrl string, ptahToken string) *Client {
	return &Client{
		BaseUrl:   baseUrl,
		ptahToken: ptahToken,
		http:      &http.Client{},
	}
}

func (c *Client) Started(ctx context.Context, req StartedReq) (*StartedRes, error) {
	var result StartedRes

	_, err := c.send(ctx, "POST", "/events/started", req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) send(ctx context.Context, method, url string, req interface{}, res interface{}) ([]byte, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, method, c.url(url), bytes.NewReader(data))
	if err != nil {

		return nil, err
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Ptah-Token", c.ptahToken)

	response, err := c.http.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		return body, json.Unmarshal(body, &res)
	}

	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("ptah error: %s", string(body))
	}

	var serviceError ServiceError
	err = json.Unmarshal(body, &serviceError)
	if err == nil {
		return body, &serviceError
	}

	return body, fmt.Errorf("unexpected status code: %d, response: %s", response.StatusCode, string(body))
}

func (c *Client) url(url string) string {
	return fmt.Sprintf("%s%s", c.BaseUrl, url)
}
