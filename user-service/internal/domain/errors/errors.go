package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUserNotFound      = status.Error(codes.NotFound, "user not found")
	ErrProfileNotFound   = status.Error(codes.NotFound, "profile not found")
	ErrInvalidInput      = status.Error(codes.InvalidArgument, "invalid input")
	ErrDatabase          = status.Error(codes.Internal, "database error")
	ErrInternalServer    = status.Error(codes.Internal, "internal server error")
	ErrUnauthorized      = status.Error(codes.Unauthenticated, "unauthorized")
	ErrForbidden         = status.Error(codes.PermissionDenied, "forbidden")
	ErrUserAlreadyExists = status.Error(codes.AlreadyExists, "user already exists")
)

func IsGRPCError(err error) bool {
	_, ok := status.FromError(err)
	return ok
}

func ToGRPCError(err error) error {
	if IsGRPCError(err) {
		return err
	}
	return ErrInternalServer
}

