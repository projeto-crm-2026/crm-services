package planservice

import "errors"

var (
	ErrInvalidTransition    = errors.New("invalid subscription status transition")
	ErrOrganizationNotFound = errors.New("organization not found in subscription records")
	ErrFreePlanNotFound     = errors.New("free plan not found in catalog")
)
