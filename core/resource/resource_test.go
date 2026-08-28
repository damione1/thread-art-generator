package resource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectKeyHelpers(t *testing.T) {
	t.Parallel()
	require.Equal(t, "users/u1/arts/a1/original", ArtOriginalObjectKey("u1", "a1"))
	require.Equal(t, "users/u1/arts/a1/compositions/c1/preview", CompositionPreviewObjectKey("u1", "a1", "c1"))
	require.Equal(t, "users/u1/arts/a1/compositions/c1/gcode", CompositionGcodeObjectKey("u1", "a1", "c1"))
	require.Equal(t, "users/u1/arts/a1/compositions/c1/pathlist", CompositionPathlistObjectKey("u1", "a1", "c1"))
}

func TestArtImageObjectKeyDualRun(t *testing.T) {
	t.Parallel()
	user, art := "user-uuid", "art-uuid"
	require.Equal(t, ArtOriginalObjectKey(user, art), ArtImageObjectKey(user, art, art))
	require.Equal(t, ArtOriginalObjectKey(user, art), ArtImageObjectKey(user, art, ""))
	require.Equal(t, BuildArtResourceName(user, "legacy-image"), ArtImageObjectKey(user, art, "legacy-image"))
}

func TestParseResourceName(t *testing.T) {
	t.Parallel()
	got, err := ParseResourceName("users/u1/arts/a1")
	require.NoError(t, err)
	art, ok := got.(*Art)
	require.True(t, ok)
	require.Equal(t, "u1", art.UserID)
	require.Equal(t, "a1", art.ArtID)

	_, err = ParseResourceName("not-a-resource")
	require.Error(t, err)
}
