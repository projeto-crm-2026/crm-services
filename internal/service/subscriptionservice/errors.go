package subscriptionservice

import "errors"

var (
	ErrAlreadySubscribed   = errors.New("organization already has an active paid subscription")
	ErrNoPaidSubscription  = errors.New("organization does not have an active paid subscription")
	ErrFreePlanNotPaid     = errors.New("cannot subscribe to the free plan via payment")
	ErrPlanNotFound        = errors.New("plan not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrSamePlan            = errors.New("organization is already on this plan")
	ErrDowngradeNotAllowed = errors.New("downgrade must be done via cancellation")
)
