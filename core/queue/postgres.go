package queue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	defaultMaxAttempts       = 5
	defaultPollInterval      = time.Second
	defaultVisibilityTimeout = 10 * time.Minute
	defaultBaseBackoff       = 10 * time.Second
)

const (
	insertJobSQL = `
INSERT INTO jobs (topic, body, status, attempts, available_at)
VALUES ($1, $2, 'pending', 0, now())`

	// Subquery FOR UPDATE SKIP LOCKED (not a PG12-inlinable CTE).
	claimJobSQL = `
UPDATE jobs
SET
    status = 'processing',
    consumer = $2,
    available_at = now() + ($3 * interval '1 millisecond'),
    updated_at = now()
WHERE id = (
    SELECT id
    FROM jobs
    WHERE topic = $1
      AND status IN ('pending', 'processing')
      AND available_at <= now()
    ORDER BY created_at
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
RETURNING id, body, attempts`

	completeJobSQL = `
UPDATE jobs
SET status = 'done', consumer = NULL, updated_at = now()
WHERE id = $1 AND consumer = $2 AND status = 'processing'`

	failJobSQL = `
UPDATE jobs
SET
    attempts = attempts + 1,
    consumer = NULL,
    status = CASE WHEN attempts + 1 >= $3 THEN 'dead' ELSE 'pending' END,
    available_at = CASE
        WHEN attempts + 1 >= $3 THEN now()
        ELSE now() + ($4 * power(2, attempts) * interval '1 millisecond')
    END,
    updated_at = now()
WHERE id = $1 AND consumer = $2 AND status = 'processing'`
)

// PostgresOptions tunes PostgresQueue. Zero values mean defaults.
type PostgresOptions struct {
	MaxAttempts       int
	PollInterval      time.Duration
	VisibilityTimeout time.Duration
	BaseBackoff       time.Duration
}

func (o PostgresOptions) withDefaults() PostgresOptions {
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = defaultMaxAttempts
	}
	if o.PollInterval <= 0 {
		o.PollInterval = defaultPollInterval
	}
	if o.VisibilityTimeout <= 0 {
		o.VisibilityTimeout = defaultVisibilityTimeout
	}
	if o.BaseBackoff <= 0 {
		o.BaseBackoff = defaultBaseBackoff
	}
	return o
}

// PostgresQueue implements Queue (and QueueClient) with FOR UPDATE SKIP LOCKED.
// Dual-run: Pub/Sub stays the worker transport until Phase C.
type PostgresQueue struct {
	db   *sql.DB
	opts PostgresOptions

	closeOnce sync.Once
	done      chan struct{}
	rootCtx   context.Context
	cancel    context.CancelFunc
}

var (
	_ Queue       = (*PostgresQueue)(nil)
	_ QueueClient = (*PostgresQueue)(nil)
)

// NewPostgresQueue wraps an existing *sql.DB. It does not close the DB.
func NewPostgresQueue(db *sql.DB, opts PostgresOptions) *PostgresQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &PostgresQueue{
		db:      db,
		opts:    opts.withDefaults(),
		done:    make(chan struct{}),
		rootCtx: ctx,
		cancel:  cancel,
	}
}

// Publish inserts a pending job.
func (q *PostgresQueue) Publish(ctx context.Context, topic string, body []byte) error {
	if err := q.errIfClosed(); err != nil {
		return err
	}
	if topic == "" {
		return fmt.Errorf("queue: topic is required")
	}
	_, err := q.db.ExecContext(ctx, insertJobSQL, topic, body)
	if err != nil {
		return fmt.Errorf("queue: insert job: %w", err)
	}
	return nil
}

// PublishMessage implements QueueClient. queueName maps 1:1 onto topic.
func (q *PostgresQueue) PublishMessage(ctx context.Context, queueName string, message []byte) error {
	return q.Publish(ctx, queueName, message)
}

// Subscribe blocks, claiming jobs for topic with SKIP LOCKED until ctx or Close.
func (q *PostgresQueue) Subscribe(ctx context.Context, topic, consumer string, h Handler) error {
	if h == nil {
		return fmt.Errorf("queue: handler is required")
	}
	if topic == "" {
		return fmt.Errorf("queue: topic is required")
	}
	if consumer == "" {
		consumer = "default"
	}
	if err := q.errIfClosed(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(q.rootCtx, cancel)
	defer stop()

	for {
		if err := ctx.Err(); err != nil {
			return q.subscribeExit(err)
		}

		job, err := q.claim(ctx, topic, consumer)
		if err != nil {
			if ctx.Err() != nil {
				return q.subscribeExit(ctx.Err())
			}
			return err
		}
		if job == nil {
			select {
			case <-ctx.Done():
				return q.subscribeExit(ctx.Err())
			case <-time.After(q.opts.PollInterval):
			}
			continue
		}

		if err := h(ctx, job.body); err != nil {
			if failErr := q.fail(ctx, job.id, consumer); failErr != nil {
				return failErr
			}
			continue
		}
		if err := q.complete(ctx, job.id, consumer); err != nil {
			return err
		}
	}
}

// Close cancels in-flight Subscribe loops. Does not close the *sql.DB.
func (q *PostgresQueue) Close() error {
	q.closeOnce.Do(func() {
		close(q.done)
		q.cancel()
	})
	return nil
}

type claimedJob struct {
	id       string
	body     []byte
	attempts int
}

func (q *PostgresQueue) claim(ctx context.Context, topic, consumer string) (*claimedJob, error) {
	visMs := q.opts.VisibilityTimeout.Milliseconds()
	var job claimedJob
	err := q.db.QueryRowContext(ctx, claimJobSQL, topic, consumer, visMs).Scan(&job.id, &job.body, &job.attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("queue: claim job: %w", err)
	}
	return &job, nil
}

func (q *PostgresQueue) complete(ctx context.Context, id, consumer string) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	res, err := q.db.ExecContext(ctx, completeJobSQL, id, consumer)
	if err != nil {
		return fmt.Errorf("queue: complete job: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		log.Warn().Str("id", id).Str("consumer", consumer).Msg("queue: complete missed (lost claim)")
	}
	return nil
}

func (q *PostgresQueue) fail(ctx context.Context, id, consumer string) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	res, err := q.db.ExecContext(ctx, failJobSQL, id, consumer, q.opts.MaxAttempts, q.opts.BaseBackoff.Milliseconds())
	if err != nil {
		return fmt.Errorf("queue: fail job: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		log.Warn().Str("id", id).Str("consumer", consumer).Msg("queue: fail missed (lost claim)")
	}
	return nil
}

func (q *PostgresQueue) errIfClosed() error {
	select {
	case <-q.done:
		return fmt.Errorf("queue: closed")
	default:
		return nil
	}
}

func (q *PostgresQueue) isClosed() bool {
	select {
	case <-q.done:
		return true
	default:
		return false
	}
}

func (q *PostgresQueue) subscribeExit(err error) error {
	if q.isClosed() {
		return nil
	}
	return err
}
