package ptah_client

import (
	"context"
	"strings"
)

func (c *Client) SendMetrics(ctx context.Context, metrics []string) error {
	body := strings.Join(metrics, "\n")

	_, err := c.send(ctx, "POST", "/metrics", body, nil)

	return err
}
