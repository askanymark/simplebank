package core

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FieldViolation returns a BadRequest_FieldViolation with the given field and error message
func FieldViolation(field string, err error) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: err.Error(),
	}
}

// InvalidArgumentError creates and returns an error representing invalid argument violations with detailed field errors
func InvalidArgumentError(violations []*errdetails.BadRequest_FieldViolation) error {
	badRequest := &errdetails.BadRequest{FieldViolations: violations}
	statusInvalid := status.New(codes.InvalidArgument, "invalid arguments")
	statusDetails, err := statusInvalid.WithDetails(badRequest)
	if err != nil {
		return statusInvalid.Err()
	}

	return statusDetails.Err()
}

// UnauthenticatedError creates a gRPC error with the 'Unauthenticated' code and includes the provided error message.
func UnauthenticatedError(err error) error {
	return status.Errorf(codes.Unauthenticated, "unauthenticated: %s", err)
}
