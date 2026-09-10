package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"velostats/internal/support"
)

type Handler func(ctx context.Context, payload json.RawMessage) error

type Worker struct {
	client   *Client
	queues   []string
	handlers map[string]Handler
	db       *sql.DB
	logger   *slog.Logger
}

func NewWorker(client *Client, queues []string, db *sql.DB, logger *slog.Logger) *Worker {
	return &Worker{
		client:   client,
		queues:   queues,
		handlers: map[string]Handler{},
		db:       db,
		logger:   logger,
	}
}

func (w *Worker) Register(jobType string, handler Handler) {
	w.handlers[jobType] = handler
}

// Jobs are run one at a time, so a worker on a single queue never calls an
// upstream API concurrently.
func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info("worker started", "queues", w.queues)

	for {
		if ctx.Err() != nil {
			w.logger.Info("worker stopped", "queues", w.queues)

			return nil
		}

		job, err := w.client.Pop(ctx, w.queues, time.Second)
		if errors.Is(err, context.Canceled) {
			continue
		}

		if err != nil {
			w.logger.Error("failed to pop job", "error", err)
			time.Sleep(time.Second)

			continue
		}

		if job == nil {
			continue
		}

		w.run(ctx, *job)
	}
}

func (w *Worker) run(ctx context.Context, job Job) {
	handler, known := w.handlers[job.Type]
	if !known {
		w.logger.Error("no handler registered for job", "job", job.Type, "id", job.ID)
		w.recordFailure(job, errors.New("no handler registered"))

		return
	}

	started := time.Now()

	if err := handler(ctx, job.Payload); err != nil {
		w.logger.Error("job failed", "job", job.Type, "id", job.ID, "error", err)
		w.recordFailure(job, err)

		return
	}

	w.logger.Info("job processed", "job", job.Type, "id", job.ID, "duration", time.Since(started).String())
}

// Jobs are run once and are not retried, matching the workers' --tries=1 behaviour.
func (w *Worker) recordFailure(job Job, cause error) {
	if _, err := w.db.Exec(
		`INSERT INTO failed_jobs (uuid, queue, job_type, payload, exception, failed_at) VALUES (?, ?, ?, ?, ?, ?)`,
		job.ID, job.Queue, job.Type, string(job.Payload), cause.Error(), support.DatabaseDateTime(time.Now()),
	); err != nil {
		w.logger.Error("failed to record failed job", "job", job.Type, "id", job.ID, "error", err)
	}
}
