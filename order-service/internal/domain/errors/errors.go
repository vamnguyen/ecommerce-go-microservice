package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrOrderNotFound     = status.Error(codes.NotFound, "order not found")
	ErrInvalidInput      = status.Error(codes.InvalidArgument, "invalid input")
	ErrDatabase          = status.Error(codes.Internal, "database error")
	ErrInternalServer    = status.Error(codes.Internal, "internal server error")
	ErrUnauthorized      = status.Error(codes.Unauthenticated, "unauthorized")
	ErrForbidden         = status.Error(codes.PermissionDenied, "forbidden")
	ErrInvalidStatus     = status.Error(codes.InvalidArgument, "invalid order status")
	ErrStatusTransition  = status.Error(codes.InvalidArgument, "invalid status transition")
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

