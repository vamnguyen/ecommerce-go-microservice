package handler

import (
	domainErr "order-service/internal/domain/errors"
)

func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	if domainErr.IsGRPCError(err) {
		return err
	}

	return domainErr.ToGRPCError(err)
}

