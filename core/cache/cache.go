package cache

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gocloud.dev/blob"
)

// Cache entry that includes the URL and its expiration time
type cacheEntry struct {
	url       string
	expiresAt time.Time
}

// Global cache and mutex to protect access
var (
	urlCache = make(map[string]cacheEntry)
	mu       sync.Mutex
)

func GetOrCreateSignedImageURL(ctx context.Context, bucket *blob.Bucket, imageID string, expiry int) (string, error) {
	// Lock the mutex to ensure exclusive access
	mu.Lock()
	defer mu.Unlock()

	// Check if the URL is in the cache
	if entry, ok := urlCache[imageID]; ok {
		if time.Now().Before(entry.expiresAt) {
			return entry.url, nil
		}
		// Remove expired entry
		delete(urlCache, imageID)
	}

	// Generate a new signed URL
	imageUrl, err := bucket.SignedURL(ctx, imageID, &blob.SignedURLOptions{Method: "GET", Expiry: time.Duration(expiry) * time.Minute})
	if err != nil {
		return "", err
	}

	// Cache the signed URL with its expiration time
	urlCache[imageID] = cacheEntry{
		url:       imageUrl,
		expiresAt: time.Now().Add(10 * time.Minute),
	}

	return imageUrl, nil
}

func CleanExpiredCacheEntries() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		<-ticker.C
		mu.Lock()
		for id, entry := range urlCache {
			if time.Now().After(entry.expiresAt) {
				delete(urlCache, id)
			}
		}
		log.Info().Msg("Cleaned expired cache entries")
		mu.Unlock()
	}
}
