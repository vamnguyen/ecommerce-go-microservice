package handler

import (
	domainErr "auth-service/internal/domain/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch err {
	case domainErr.ErrUserNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domainErr.ErrUserAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case domainErr.ErrInvalidCredentials:
		return status.Error(codes.Unauthenticated, err.Error())
	case domainErr.ErrAccountLocked, domainErr.ErrAccountInactive:
		return status.Error(codes.PermissionDenied, err.Error())
	case domainErr.ErrInvalidToken, domainErr.ErrTokenExpired, domainErr.ErrTokenRevoked, domainErr.ErrMissingToken:
		return status.Error(codes.Unauthenticated, err.Error())
	case domainErr.ErrWeakPassword, domainErr.ErrInvalidPassword:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "an internal error occurred")
	}
}
