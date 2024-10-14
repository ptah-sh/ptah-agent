package ptah_client

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
)

type GetNextTaskRes struct {
	taskReq

	TaskType int    `json:"type"`
	Payload  string `json:"payload"`
}

func (c *Client) GetNextTask(ctx context.Context) (*GetNextTaskRes, error) {
	var result GetNextTaskRes

	body, err := c.send(ctx, "GET", "/tasks/next", nil, &result)
	if err != nil {
		return nil, errors.Wrapf(err, "GET /tasks/next failed")
	}

	if body == nil {
		return nil, nil
	}

	return &result, nil
}

func (c *Client) CompleteTask(ctx context.Context, taskID int, taskResult interface{}) error {
	_, err := c.send(ctx, "POST", fmt.Sprintf("/tasks/%d/complete", taskID), taskResult, nil)
	if err != nil {
		if errors.Is(err, &HttpConflictError{}) {
			log.Printf("task %d already completed", taskID)

			return nil
		}

		return errors.Wrapf(err, "POST /tasks/%d/complete failed", taskID)
	}

	return err
}

func (c *Client) FailTask(ctx context.Context, taskID int, result interface{}) error {
	_, err := c.send(ctx, "POST", fmt.Sprintf("/tasks/%d/fail", taskID), result, nil)
	if err != nil {
		if errors.Is(err, &HttpConflictError{}) {
			log.Printf("task %d already failed", taskID)

			return nil
		}

		return errors.Wrapf(err, "POST /tasks/%d/fail failed", taskID)
	}

	return err
}
