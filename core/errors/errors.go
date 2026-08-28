package pbErrors

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
	"github.com/rs/zerolog/log"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

const (
	ErrEmailAlreadyExists        = "email already exists"
	ErrInvalidResourceName       = "invalid resource name"
	ErrUserNotFound              = "user not found"
	ErrIncorrectCredentials      = "incorrect email or password"
	ErrUserNotActive             = "user is not active"
	ErrTooManyValidationRequests = "too many validation requests"
	ErrSessionNotFound           = "session not found"
	ErrSessionBlocked            = "session is blocked"

	// UserFacingInternalMessage is what forms show for CodeInternal.
	// The wire message stays the operator-safe first argument to InternalError.
	UserFacingInternalMessage = "Something went wrong. Please try again."
)

// FieldViolation creates a google.rpc.BadRequest field violation.
// field is the proto path (snake_case, dotted), e.g. "art.title".
func FieldViolation(field string, err error) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: err.Error(),
	}
}

func withBadRequest(code connect.Code, message string, violations []*errdetails.BadRequest_FieldViolation) error {
	cerr := connect.NewError(code, errors.New(message))
	if len(violations) == 0 {
		return cerr
	}
	detail, err := connect.NewErrorDetail(&errdetails.BadRequest{FieldViolations: violations})
	if err != nil {
		return cerr
	}
	cerr.AddDetail(detail)
	return cerr
}

func InvalidArgumentError(violations []*errdetails.BadRequest_FieldViolation) error {
	return withBadRequest(connect.CodeInvalidArgument, "invalid parameters", violations)
}

func UnauthenticatedError(message string) error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New(message))
}

func PermissionDeniedError(message string) error {
	return connect.NewError(connect.CodePermissionDenied, errors.New(message))
}

// InternalError returns CodeInternal with message only. cause is logged, never put on the wire.
func InternalError(message string, cause error) error {
	if cause != nil {
		log.Error().Err(cause).Str("public", message).Msg("internal error")
	}
	return connect.NewError(connect.CodeInternal, errors.New(message))
}

func NotFoundError(message string) error {
	return connect.NewError(connect.CodeNotFound, errors.New(message))
}

func AlreadyExistsError(message string, field string) error {
	if field == "" {
		return connect.NewError(connect.CodeAlreadyExists, errors.New(message))
	}
	return withBadRequest(connect.CodeAlreadyExists, message, []*errdetails.BadRequest_FieldViolation{
		FieldViolation(field, errors.New(message)),
	})
}

func FailedPreconditionError(message string) error {
	return connect.NewError(connect.CodeFailedPrecondition, errors.New(message))
}

func hasCode(err error, code connect.Code) bool {
	return err != nil && connect.CodeOf(err) == code
}

func IsNotFoundError(err error) bool {
	return hasCode(err, connect.CodeNotFound)
}

func IsUnauthenticatedError(err error) bool {
	return hasCode(err, connect.CodeUnauthenticated)
}

func IsPermissionDeniedError(err error) bool {
	return hasCode(err, connect.CodePermissionDenied)
}

func IsInternalError(err error) bool {
	return hasCode(err, connect.CodeInternal)
}

func IsInvalidArgumentError(err error) bool {
	return hasCode(err, connect.CodeInvalidArgument)
}

func ExtractFieldViolations(err error) []*errdetails.BadRequest_FieldViolation {
	if err == nil {
		return nil
	}
	var ce *connect.Error
	if !errors.As(err, &ce) {
		return nil
	}
	for _, d := range ce.Details() {
		msg, valErr := d.Value()
		if valErr != nil {
			continue
		}
		if badRequest, ok := msg.(*errdetails.BadRequest); ok {
			return badRequest.GetFieldViolations()
		}
	}
	return nil
}

func HasFieldViolation(err error, field string) bool {
	return GetFieldViolationMessage(err, field) != ""
}

func GetFieldViolationMessage(err error, field string) string {
	for _, v := range ExtractFieldViolations(err) {
		if v.GetField() == field {
			return v.GetDescription()
		}
	}
	return ""
}

// ConvertProtoValidateError turns protovalidate violations into InvalidArgument
// with proto field paths (art.title, composition.nails_quantity, page_size).
func ConvertProtoValidateError(err error) error {
	if err == nil {
		return nil
	}

	var validationErr *protovalidate.ValidationError
	if !errors.As(err, &validationErr) {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("validation failed"))
	}

	fieldViolations := make([]*errdetails.BadRequest_FieldViolation, 0, len(validationErr.Violations))
	for _, violation := range validationErr.Violations {
		fieldPath := ""
		if violation != nil && violation.Proto != nil {
			fieldPath = protovalidate.FieldPathString(violation.Proto.GetField())
		}
		message := "validation failed"
		if violation != nil && violation.Proto != nil && violation.Proto.GetMessage() != "" {
			message = violation.Proto.GetMessage()
		}
		fieldViolations = append(fieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       fieldPath,
			Description: message,
		})
	}

	return InvalidArgumentError(fieldViolations)
}
