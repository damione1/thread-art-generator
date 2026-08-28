package storage

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

type memoryObject struct {
	data        []byte
	contentType string
}

// MemoryBucket is an in-process Bucket for tests. No network.
type MemoryBucket struct {
	mu      sync.Mutex
	objects map[string]memoryObject
}

func NewMemoryBucket() *MemoryBucket {
	return &MemoryBucket{objects: make(map[string]memoryObject)}
}

func (m *MemoryBucket) Put(_ context.Context, key string, r io.Reader, opts PutOptions) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	ct := opts.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}
	m.objects[key] = memoryObject{data: data, contentType: ct}
	return nil
}

func (m *MemoryBucket) Get(_ context.Context, key string) (io.ReadCloser, *ObjectInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	obj, ok := m.objects[key]
	if !ok {
		return nil, nil, ErrObjectNotFound
	}
	return io.NopCloser(bytes.NewReader(obj.data)), m.info(key, obj), nil
}

func (m *MemoryBucket) Head(_ context.Context, key string) (*ObjectInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	obj, ok := m.objects[key]
	if !ok {
		return nil, ErrObjectNotFound
	}
	return m.info(key, obj), nil
}

func (m *MemoryBucket) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.objects[key]; !ok {
		return ErrObjectNotFound
	}
	delete(m.objects, key)
	return nil
}

func (m *MemoryBucket) PresignPut(_ context.Context, key string, opts PresignPutOptions) (*PresignPut, error) {
	ttl := opts.TTL
	if ttl == 0 {
		ttl = 10 * time.Minute
	}
	headers := http.Header{}
	if opts.ContentType != "" {
		headers.Set("Content-Type", opts.ContentType)
	}
	return &PresignPut{
		URL:     "http://memory.local/put/" + key,
		Method:  http.MethodPut,
		Headers: headers,
		Expires: time.Now().Add(ttl),
	}, nil
}

func (m *MemoryBucket) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return "http://memory.local/get/" + key, nil
}

func (m *MemoryBucket) info(key string, obj memoryObject) *ObjectInfo {
	return &ObjectInfo{
		Key:         key,
		Size:        int64(len(obj.data)),
		ContentType: obj.contentType,
		ETag:        "memory",
	}
}

var _ Bucket = (*MemoryBucket)(nil)
