package pbErrors

import (
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/bufbuild/protovalidate-go"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func TestFailedPreconditionError(t *testing.T) {
	t.Parallel()
	err := FailedPreconditionError("uploaded object exceeds 10MB")
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "exceeds 10MB")
	got := FromConnectError(err)
	require.Equal(t, ErrorTypePrecondition, got.Type)
	require.Equal(t, "uploaded object exceeds 10MB", got.GlobalError)
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

func TestFromConnectErrorExtractsFieldViolations(t *testing.T) {
	t.Parallel()
	err := InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
		FieldViolation("art.title", errors.New("Title is required")),
	})
	got := FromConnectError(err)
	require.Equal(t, ErrorTypeValidation, got.Type)
	require.Equal(t, []string{"Title is required"}, got.Fields["art.title"])
	require.Empty(t, got.GlobalError)
}

func TestFromConnectErrorUnwraps(t *testing.T) {
	t.Parallel()
	inner := InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
		FieldViolation("art.title", errors.New("Title is required")),
	})
	got := FromConnectError(fmt.Errorf("bff: %w", inner))
	require.Equal(t, []string{"Title is required"}, got.Fields["art.title"])
}

func TestInternalErrorDoesNotLeakCause(t *testing.T) {
	t.Parallel()
	err := InternalError("failed to insert art", errors.New("pq: duplicate key value"))
	require.True(t, IsInternalError(err))
	require.Contains(t, err.Error(), "failed to insert art")
	require.NotContains(t, err.Error(), "duplicate key")

	var ce *connect.Error
	require.True(t, errors.As(err, &ce))
	require.Empty(t, ce.Details())

	got := FromConnectError(err)
	require.Equal(t, ErrorTypeInternal, got.Type)
	require.Equal(t, UserFacingInternalMessage, got.GlobalError)
	require.Equal(t, "failed to insert art", got.Message)

	form := FormFields(err)
	require.Equal(t, []string{UserFacingInternalMessage}, form["_form"])
}

func TestConvertProtoValidateErrorArtTitle(t *testing.T) {
	t.Parallel()
	err := protovalidate.Validate(&pb.CreateArtRequest{
		Parent: "users/abc",
		Art:    &pb.Art{Title: ""},
	})
	require.Error(t, err)

	converted := ConvertProtoValidateError(err)
	require.True(t, IsInvalidArgumentError(converted))
	require.True(t, HasFieldViolation(converted, "art.title"), "got violations: %v", ExtractFieldViolations(converted))
	require.Equal(t, "Title is required", GetFieldViolationMessage(converted, "art.title"))

	form := FormFields(converted)
	require.Equal(t, []string{"Title is required"}, FieldMessages(form, "art.title"))
	require.Equal(t, []string{"Title is required"}, FieldMessages(form, "title"))
}

func TestConvertProtoValidateErrorCompositionNails(t *testing.T) {
	t.Parallel()
	err := protovalidate.Validate(&pb.CreateCompositionRequest{
		Parent: "users/abc/arts/def",
		Composition: &pb.Composition{
			NailsQuantity:     0,
			ImgSize:           800,
			MaxPaths:          10000,
			StartingNail:      0,
			MinimumDifference: 10,
			BrightnessFactor:  50,
			ImageContrast:     40,
			PhysicalRadius:    609.6,
		},
	})
	require.Error(t, err)

	converted := ConvertProtoValidateError(err)
	require.True(t, HasFieldViolation(converted, "composition.nails_quantity"), "got violations: %v", ExtractFieldViolations(converted))

	form := FormFields(converted)
	require.NotEmpty(t, FieldMessages(form, "composition.nails_quantity"))
	require.NotEmpty(t, FieldMessages(form, "nails_quantity"))
}

func TestFieldMessagesAliases(t *testing.T) {
	t.Parallel()
	fields := map[string][]string{
		"art.title":                  {"Title is required"},
		"composition.nails_quantity": {"must be > 0"},
		"page_size":                  {"must be <= 100"},
	}
	expanded := ExpandFieldKeys(fields)
	require.Equal(t, []string{"Title is required"}, FieldMessages(expanded, "title"))
	require.Equal(t, []string{"Title is required"}, FieldMessages(expanded, "art.title"))
	require.Equal(t, []string{"must be > 0"}, FieldMessages(expanded, "nails_quantity"))
	require.Equal(t, []string{"must be > 0"}, FieldMessages(expanded, "nailsQuantity"))
	require.Equal(t, []string{"must be <= 100"}, FieldMessages(expanded, "pageSize"))
	require.Equal(t, []string{"must be <= 100"}, FieldMessages(expanded, "page_size"))
}

func TestFormFieldsValidationDoesNotPromoteMessageToForm(t *testing.T) {
	t.Parallel()
	err := InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
		FieldViolation("art.title", errors.New("Title is required")),
	})
	form := FormFields(err)
	require.Equal(t, []string{"Title is required"}, form["art.title"])
	require.Empty(t, form["_form"])
}

func TestToConnectErrorRoundTrip(t *testing.T) {
	t.Parallel()
	original := InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
		FieldViolation("email", errors.New("email already exists")),
	})
	decoded := FromConnectError(original)
	again := FromConnectError(decoded.ToConnectError())
	require.Equal(t, decoded.Fields["email"], again.Fields["email"])
}
