// Package queue pushes background jobs onto Redis and runs the workers that
// drain them.
package queue

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// The queues the application dispatches onto. The two check queues are each
// drained by a single worker, so the free upstream APIs are never called
// concurrently.
const (
	Default            = "default"
	RideDistanceChecks = "ride_distance_checks"
	RideWeatherChecks  = "ride_weather_checks"
)

// Job is one unit of background work as it travels through Redis.
type Job struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Queue    string          `json:"queue"`
	Payload  json.RawMessage `json:"payload"`
	PushedAt time.Time       `json:"pushed_at"`
}

// Client pushes and pops jobs on Redis.
type Client struct {
	redis  *redis.Client
	prefix string
}

// NewClient builds a queue client over the given Redis connection. Keys are
// namespaced with the prefix so several applications can share one Redis.
func NewClient(client *redis.Client, prefix string) *Client {
	return &Client{redis: client, prefix: prefix}
}

// Close releases the Redis connection.
func (c *Client) Close() error {
	return c.redis.Close()
}

// Ping checks that Redis is reachable.
func (c *Client) Ping(ctx context.Context) error {
	return c.redis.Ping(ctx).Err()
}

// Push adds a job to the end of a queue.
func (c *Client) Push(ctx context.Context, queue, jobType string, payload any) error {
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode payload for %s: %w", jobType, err)
	}

	job := Job{
		ID:       newJobID(),
		Type:     jobType,
		Queue:    queue,
		Payload:  encodedPayload,
		PushedAt: time.Now().UTC(),
	}

	encodedJob, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("encode job %s: %w", jobType, err)
	}

	if err := c.redis.RPush(ctx, c.key(queue), encodedJob).Err(); err != nil {
		return fmt.Errorf("push job %s onto %s: %w", jobType, queue, err)
	}

	return nil
}

// Pop takes the next job off any of the given queues, waiting up to the timeout
// for one to arrive. It returns nil when the wait times out.
func (c *Client) Pop(ctx context.Context, queues []string, timeout time.Duration) (*Job, error) {
	keys := make([]string, 0, len(queues))
	for _, queue := range queues {
		keys = append(keys, c.key(queue))
	}

	popped, err := c.redis.BLPop(ctx, timeout, keys...).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("pop job: %w", err)
	}

	if len(popped) != 2 {
		return nil, fmt.Errorf("pop job: unexpected reply of %d values", len(popped))
	}

	var job Job
	if err := json.Unmarshal([]byte(popped[1]), &job); err != nil {
		return nil, fmt.Errorf("decode job: %w", err)
	}

	return &job, nil
}

// Size reports how many jobs are waiting on a queue.
func (c *Client) Size(ctx context.Context, queue string) (int64, error) {
	size, err := c.redis.LLen(ctx, c.key(queue)).Result()
	if err != nil {
		return 0, fmt.Errorf("measure queue %s: %w", queue, err)
	}

	return size, nil
}

func (c *Client) key(queue string) string {
	return c.prefix + queue
}

func newJobID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(id)
}
