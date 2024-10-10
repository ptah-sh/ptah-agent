package ptah_agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	ptahClient "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

type SafeClient struct {
	client *ptahClient.Client
	db     *sql.DB
}

func NewSafeClient(client *ptahClient.Client, ptahRootDir string) (*SafeClient, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s/ptah-agent.db", ptahRootDir))
	if err != nil {
		return nil, err
	}

	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA synchronous=normal;")
	db.Exec("PRAGMA temp_store=memory;")
	db.Exec("PRAGMA cache_size=1000000;")

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='requests';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		_, err = db.Exec("CREATE TABLE requests (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, payload TEXT);")
		if err != nil {
			return nil, err
		}
	}

	return &SafeClient{client: client, db: db}, nil
}

func (c *SafeClient) CompleteTask(ctx context.Context, taskID int, result interface{}) error {
	jsonResult, err := json.Marshal(struct {
		TaskID int
		Result interface{}
	}{
		TaskID: taskID,
		Result: result,
	})

	if err != nil {
		return err
	}

	_, err = c.db.Exec("INSERT INTO requests (name, payload) VALUES (?, ?)", "CompleteTask", jsonResult)
	if err != nil {
		return err
	}

	return nil
}

func (c *SafeClient) FailTask(ctx context.Context, taskID int, taskError *ptahClient.TaskError) error {
	jsonTaskError, err := json.Marshal(struct {
		TaskID    int
		TaskError *ptahClient.TaskError
	}{
		TaskID:    taskID,
		TaskError: taskError,
	})
	if err != nil {
		return err
	}

	_, err = c.db.Exec("INSERT INTO requests (name, payload) VALUES (?, ?)", "FailTask", jsonTaskError)
	if err != nil {
		return err
	}

	return nil
}

func (c *SafeClient) SendMetrics(ctx context.Context, metrics []string) error {
	jsonMetrics, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	_, err = c.db.Exec("INSERT INTO requests (name, payload) VALUES (?, ?)", "SendMetrics", jsonMetrics)
	if err != nil {
		return err
	}

	return nil
}

// TODO: split into "sequential" and "parallel". One for tasks, another for metrics. Metrics should be sent always in parallel, in background.
func (c *SafeClient) PerformForegroundRequests(ctx context.Context) error {
	rows, err := c.db.Query("SELECT id, name, payload FROM requests WHERE name in ('CompleteTask', 'FailTask');")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var rowPayload []byte

		err = rows.Scan(&id, &name, &rowPayload)
		if err != nil {
			return err
		}

		switch name {
		case "CompleteTask":
			var payload struct {
				TaskID int
				Result interface{}
			}

			err = json.Unmarshal(rowPayload, &payload)
			if err != nil {
				return err
			}

			// TODO: if the http status is 409 (Conflict) - it is ok, than the task result is already saved and we are executing the same task again due to some error/crash
			err = c.client.CompleteTask(ctx, payload.TaskID, payload.Result)
			if err != nil {
				return err
			}
		case "FailTask":
			var payload struct {
				TaskID    int
				TaskError *ptahClient.TaskError
			}

			err = json.Unmarshal(rowPayload, &payload)
			if err != nil {
				return err
			}

			// TODO: if the http status is 409 (Conflict) - it is ok, than the task result is already saved and we are executing the same task again due to some error/crash
			err = c.client.FailTask(ctx, payload.TaskID, payload.TaskError)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown request name: %s", name)
		}

		_, err = c.db.Exec("DELETE FROM requests WHERE id = ?", id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *SafeClient) PerformBackgroundRequests(ctx context.Context) error {
	rows, err := c.db.Query("SELECT id, name, payload FROM requests WHERE name in ('SendMetrics');")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var rowPayload []byte

		err = rows.Scan(&id, &name, &rowPayload)
		if err != nil {
			return err
		}

		var metrics []string
		err = json.Unmarshal(rowPayload, &metrics)
		if err != nil {
			return err
		}

		err = c.client.SendMetrics(ctx, metrics)
		if err != nil {
			return err
		}

		_, err = c.db.Exec("DELETE FROM requests WHERE id = ?", id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *SafeClient) StartBackgroundRequestsProcessing(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := c.PerformBackgroundRequests(ctx)
				if err != nil {
					log.Println("error performing background requests:", err)
				}
			}
		}
	}()

	return nil
}

func (c *SafeClient) Close() error {
	return c.db.Close()
}
