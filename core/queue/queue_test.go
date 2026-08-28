package queue

import (
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

func TestClaimSQLUsesSkipLocked(t *testing.T) {
	require.Contains(t, claimJobSQL, "FOR UPDATE SKIP LOCKED")
	require.Contains(t, claimJobSQL, "status IN ('pending', 'processing')")
	require.Contains(t, insertJobSQL, "pending")
	require.Contains(t, completeJobSQL, "'done'")
	require.Contains(t, failJobSQL, "'dead'")
}
