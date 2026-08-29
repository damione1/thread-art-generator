//go:build integration

package queue

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Live SKIP LOCKED tests. Default `go test ./core/queue` skips this file.
//
//	DATABASE_URL=postgres://... go test -tags integration ./core/queue
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("skipping postgres queue test: set DATABASE_URL (and run migration 000013)")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.Ping())

	var exists bool
	err = db.QueryRow(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'jobs'
	)`).Scan(&exists)
	require.NoError(t, err)
	if !exists {
		t.Skip("jobs table missing — run migration 000013_jobs")
	}
	return db
}

func TestPostgresQueue_PublishSubscribeDone(t *testing.T) {
	db := openTestDB(t)
	topic := "test-" + uuid.NewString()
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM jobs WHERE topic = $1`, topic)
	})

	q := NewPostgresQueue(db, PostgresOptions{
		PollInterval:      20 * time.Millisecond,
		VisibilityTimeout: time.Minute,
		MaxAttempts:       3,
	})
	defer q.Close()

	body := []byte(`{"type":"composition_processing","art_id":"a","composition_id":"c"}`)
	require.NoError(t, q.Publish(t.Context(), topic, body))

	got := make(chan []byte, 1)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := q.Subscribe(ctx, topic, "test-worker", func(_ context.Context, b []byte) error {
			got <- append([]byte(nil), b...)
			return nil
		})
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Subscribe: %v", err)
		}
	}()

	select {
	case b := <-got:
		require.Equal(t, body, b)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for job")
	}
	cancel()
	wg.Wait()

	var status string
	err := db.QueryRow(`SELECT status FROM jobs WHERE topic = $1`, topic).Scan(&status)
	require.NoError(t, err)
	require.Equal(t, "done", status)
}

func TestPostgresQueue_DeadAfterMaxAttempts(t *testing.T) {
	db := openTestDB(t)
	topic := "test-" + uuid.NewString()
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM jobs WHERE topic = $1`, topic)
	})

	q := NewPostgresQueue(db, PostgresOptions{
		PollInterval:      20 * time.Millisecond,
		VisibilityTimeout: time.Minute,
		MaxAttempts:       2,
		BaseBackoff:       time.Millisecond,
	})
	defer q.Close()

	require.NoError(t, q.Publish(t.Context(), topic, []byte("nope")))

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	go func() {
		for {
			var status string
			_ = db.QueryRow(`SELECT status FROM jobs WHERE topic = $1`, topic).Scan(&status)
			if status == "dead" {
				cancel()
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(20 * time.Millisecond):
			}
		}
	}()

	err := q.Subscribe(ctx, topic, "fail-worker", func(context.Context, []byte) error {
		return errors.New("boom")
	})
	require.True(t, err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))

	var status string
	var attempts int
	err = db.QueryRow(`SELECT status, attempts FROM jobs WHERE topic = $1`, topic).Scan(&status, &attempts)
	require.NoError(t, err)
	require.Equal(t, "dead", status)
	require.Equal(t, 2, attempts)
}

func TestPostgresQueue_CloseCancelsSubscribe(t *testing.T) {
	db := openTestDB(t)
	q := NewPostgresQueue(db, PostgresOptions{PollInterval: 20 * time.Millisecond})

	errCh := make(chan error, 1)
	go func() {
		errCh <- q.Subscribe(context.Background(), "none-"+uuid.NewString(), "c", func(context.Context, []byte) error {
			return nil
		})
	}()

	time.Sleep(50 * time.Millisecond)
	require.NoError(t, q.Close())

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Subscribe did not return after Close")
	}
}
