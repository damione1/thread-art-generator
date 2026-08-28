package pbErrors

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func TestFailedPreconditionError(t *testing.T) {
	t.Parallel()
	err := FailedPreconditionError("uploaded object exceeds 10MB")
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "exceeds 10MB")
}

func TestPermissionDeniedError(t *testing.T) {
	t.Parallel()
	err := PermissionDeniedError("only the author can upload")
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
}

func TestInvalidArgumentFieldViolations(t *testing.T) {
	t.Parallel()
	err := InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
		FieldViolation("parent", errors.New("invalid resource name")),
	})
	require.True(t, IsInvalidArgumentError(err))
	require.True(t, HasFieldViolation(err, "parent"))
	require.Equal(t, "invalid resource name", GetFieldViolationMessage(err, "parent"))
}
