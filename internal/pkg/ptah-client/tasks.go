package ptah_client

import (
	"context"
	"fmt"
)

type GetNextTaskRes struct {
	task
	TaskType int    `json:"type"`
	Payload  []byte `json:"-"`
}

func (c *Client) GetNextTask(ctx context.Context) (*GetNextTaskRes, error) {
	var result GetNextTaskRes

	body, err := c.send(ctx, "GET", "/tasks/next", nil, &result)
	if err != nil {
		return nil, err
	}

	if body == nil {
		return nil, nil
	}

	result.Payload = body

	return &result, nil
}

func (c *Client) CompleteTask(ctx context.Context, taskID int, result interface{}) error {
	_, err := c.send(ctx, "POST", fmt.Sprintf("/tasks/%d/complete", taskID), result, nil)

	return err
}

func (c *Client) FailTask(ctx context.Context, taskID int, result interface{}) error {
	_, err := c.send(ctx, "POST", fmt.Sprintf("/tasks/%d/fail", taskID), result, nil)

	return err
}
