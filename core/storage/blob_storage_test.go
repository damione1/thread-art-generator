package storage

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type stubBucket struct {
	headInfo *ObjectInfo
	headErr  error
}

func (s *stubBucket) Put(context.Context, string, io.Reader, PutOptions) error { return nil }
func (s *stubBucket) Get(context.Context, string) (io.ReadCloser, *ObjectInfo, error) {
	return nil, nil, nil
}
func (s *stubBucket) Head(context.Context, string) (*ObjectInfo, error) {
	return s.headInfo, s.headErr
}
func (s *stubBucket) Delete(context.Context, string) error { return nil }
func (s *stubBucket) PresignPut(context.Context, string, PresignPutOptions) (*PresignPut, error) {
	return nil, nil
}
func (s *stubBucket) PresignGet(context.Context, string, time.Duration) (string, error) {
	return "", nil
}

func TestNewBlobStorageRejectsGCS(t *testing.T) {
	t.Parallel()
	_, err := NewBlobStorage(context.Background(), BlobStorageConfig{Provider: ProviderGCS, Bucket: "x"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "GCS is not supported")
}

func TestNewBlobStorageRejectsUnknownProvider(t *testing.T) {
	t.Parallel()
	_, err := NewBlobStorage(context.Background(), BlobStorageConfig{Provider: "firebase", Bucket: "x"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported storage provider")
}

func TestBlobStorageExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	missing := &BlobStorage{bucket: &stubBucket{headErr: ErrObjectNotFound}}
	ok, err := missing.Exists(ctx, "k")
	require.NoError(t, err)
	require.False(t, ok)

	present := &BlobStorage{bucket: &stubBucket{headInfo: &ObjectInfo{Size: 9}}}
	ok, err = present.Exists(ctx, "k")
	require.NoError(t, err)
	require.True(t, ok)
}
