package queue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompositionProcessingMessageJSON(t *testing.T) {
	msg := NewCompositionProcessingMessage("art-1", "comp-2")
	require.Equal(t, MessageTypeCompositionProcessing, msg.Type)

	raw, err := msg.ToJSON()
	require.NoError(t, err)
	require.Contains(t, string(raw), `"type":"composition_processing"`)
	require.Contains(t, string(raw), `"art_id":"art-1"`)
	require.Contains(t, string(raw), `"composition_id":"comp-2"`)

	var got CompositionProcessingMessage
	require.NoError(t, got.FromJSON(raw))
	require.Equal(t, msg.Type, got.Type)
	require.Equal(t, msg.ArtID, got.ArtID)
	require.Equal(t, msg.CompositionID, got.CompositionID)
}

func TestFromJSONRejectsGarbage(t *testing.T) {
	var msg CompositionProcessingMessage
	require.Error(t, msg.FromJSON([]byte(`{not json`)))
}

func TestNewQueueClientUnknownProvider(t *testing.T) {
	_, err := NewQueueClient(t.Context(), "mystery", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported queue provider")
}

func TestNewQueueClientRabbitMQWrongConfig(t *testing.T) {
	_, err := NewQueueClient(t.Context(), "rabbitmq", 123)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected string URL")
}

func TestPostgresQueue_PublishRequiresTopic(t *testing.T) {
	q := NewPostgresQueue(nil, PostgresOptions{})
	err := q.Publish(t.Context(), "", []byte("x"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "topic")
}

func TestPostgresQueue_SubscribeRequiresTopic(t *testing.T) {
	q := NewPostgresQueue(nil, PostgresOptions{})
	err := q.Subscribe(t.Context(), "", "w", func(context.Context, []byte) error { return nil })
	require.Error(t, err)
	require.Contains(t, err.Error(), "topic")
}

func TestPostgresOptionsDefaults(t *testing.T) {
	got := PostgresOptions{}.withDefaults()
	require.Equal(t, defaultMaxAttempts, got.MaxAttempts)
	require.Equal(t, defaultPollInterval, got.PollInterval)
	require.Equal(t, defaultVisibilityTimeout, got.VisibilityTimeout)
	require.Equal(t, defaultBaseBackoff, got.BaseBackoff)
}

func TestClaimSQLUsesSkipLocked(t *testing.T) {
	require.Contains(t, claimJobSQL, "FOR UPDATE SKIP LOCKED")
	require.Contains(t, claimJobSQL, "status IN ('pending', 'processing')")
	require.Contains(t, insertJobSQL, "pending")
	require.Contains(t, completeJobSQL, "'done'")
	require.Contains(t, failJobSQL, "'dead'")
}
