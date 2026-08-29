package pbx

import (
	"testing"

	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestPublicURL(t *testing.T) {
	t.Parallel()
	require.Equal(t, "http://localhost:9000/thread-art/users/u/arts/a/original", PublicURL("http://localhost:9000/thread-art/", "/users/u/arts/a/original"))
	require.Empty(t, PublicURL("", "key"))
	require.Empty(t, PublicURL("http://x", ""))
}

func TestArtDbToProtoNoIO(t *testing.T) {
	t.Parallel()
	art := &models.Art{
		ID:       "art-1",
		AuthorID: "user-1",
		Title:    "Sky",
		Status:   models.ArtStatusEnumCOMPLETE,
		ImageID:  null.StringFrom("art-1"),
	}
	got := ArtDbToProto(art, "")
	require.Equal(t, "users/user-1/arts/art-1", got.Name)
	require.Equal(t, "users/user-1", got.Author)
	require.Equal(t, pb.ArtStatus_ART_STATUS_COMPLETE, got.Status)
	require.Equal(t, "users/user-1/arts/art-1/original", got.ImageUrl)
}
