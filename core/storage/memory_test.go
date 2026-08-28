package storage

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryBucketRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := NewMemoryBucket()
	require.ErrorIs(t, func() error { _, err := b.Head(ctx, "missing"); return err }(), ErrObjectNotFound)

	require.NoError(t, b.Put(ctx, "k", bytes.NewReader([]byte("hi")), PutOptions{ContentType: "text/plain"}))
	info, err := b.Head(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, int64(2), info.Size)
	require.Equal(t, "text/plain", info.ContentType)

	r, info, err := b.Get(ctx, "k")
	require.NoError(t, err)
	defer r.Close()
	body, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, "hi", string(body))
	require.Equal(t, int64(2), info.Size)

	presign, err := b.PresignPut(ctx, "k", PresignPutOptions{ContentType: "text/plain"})
	require.NoError(t, err)
	require.Equal(t, http.MethodPut, presign.Method)
	require.Equal(t, "text/plain", presign.Headers.Get("Content-Type"))
}
