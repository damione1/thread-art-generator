package queue

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostgresQueue_PublishAfterClose(t *testing.T) {
	q := NewPostgresQueue(nil, PostgresOptions{})
	require.NoError(t, q.Close())
	err := q.Publish(t.Context(), TopicCompositionProcessing, []byte("x"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "closed")
}

func TestPostgresQueue_SubscribeRequiresHandler(t *testing.T) {
	q := NewPostgresQueue(nil, PostgresOptions{})
	err := q.Subscribe(t.Context(), TopicCompositionProcessing, "w", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "handler")
}
