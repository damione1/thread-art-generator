package storage

import "errors"

// Storage-related errors
var (
	ErrUnsupportedProvider = errors.New("unsupported storage provider")
	ErrObjectNotFound      = errors.New("object not found")
	ErrAccessDenied        = errors.New("access denied")
	ErrInvalidKey          = errors.New("invalid key")
	ErrInvalidContentType  = errors.New("invalid content type")
	ErrFileSizeExceeded    = errors.New("file size exceeded")
)