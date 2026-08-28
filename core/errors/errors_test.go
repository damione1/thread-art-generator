package pbErrors

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFailedPreconditionError(t *testing.T) {
	t.Parallel()
	err := FailedPreconditionError("uploaded object exceeds 10MB")
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
	require.Contains(t, err.Error(), "exceeds 10MB")
}

func TestPermissionDeniedError(t *testing.T) {
	t.Parallel()
	err := PermissionDeniedError("only the author can upload")
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}
