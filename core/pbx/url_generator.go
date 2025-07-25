package pbx

import (
	"context"

	"github.com/Damione1/thread-art-generator/core/storage"
)

// GenerateImageURL generates an image URL based on the given options
// This is a helper function that chooses between public and signed URLs
func GenerateImageURL(ctx context.Context, generator storage.ImageURLGenerator, key string, opts *storage.URLGenerationOptions) string {
	return storage.GenerateImageURL(ctx, generator, key, opts)
}
