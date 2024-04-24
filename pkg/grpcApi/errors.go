package grpcApi

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func fieldViolation(field string, err error) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: err.Error(),
	}
}

func invalidArgumentError(violations []*errdetails.BadRequest_FieldViolation) error {
	badRequest := &errdetails.BadRequest{FieldViolations: violations}
	statusInvalid := status.New(codes.InvalidArgument, "invalid parameters")

	statusDetails, err := statusInvalid.WithDetails(badRequest)
	if err != nil {
		return statusInvalid.Err()
	}

	return statusDetails.Err()
}

func unauthenticatedError(err error) error {
	return status.Errorf(codes.Unauthenticated, "unauthorized: %s", err)
}

func rolePermissionError(err error) error {
	return status.Errorf(codes.PermissionDenied, "role permission error: %s", err)
}

func internalError(err error) error {
	return status.Errorf(codes.Internal, "internal error: %s", err)
}

func notFoundError(err error) error {
	return status.Errorf(codes.NotFound, "not found: %s", err)
}
