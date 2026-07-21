package entity

import "errors"

var (
	ErrUnauthorized                     = errors.New("unauthorized")
	ErrActiveSubscription               = errors.New("active subscription is required")
	ErrNoActiveSubscription             = errors.New("you do not have an active subscription")
	ErrForbiddenTier                    = errors.New("subscription tier is not allowed to access this video")
	ErrVideoNotFound                    = errors.New("video not found")
	ErrUserNotFound                     = errors.New("user not found")
	ErrTransactionNotFound              = errors.New("transaction not found")
	ErrInvalidRequest                   = errors.New("invalid request payload")
	ErrInvalidSubscription              = errors.New("invalid subscription")
	ErrDuplicateTransaction             = errors.New("duplicate transaction")
	ErrTransactionMismatch              = errors.New("transaction payload does not match reserved transaction")
	ErrTransactionProcessed             = errors.New("transaction has already been processed")
	ErrInvalidCredentials               = errors.New("invalid email or password")
	ErrDowngradeTier                    = errors.New("active subscription cannot be downgraded")
	ErrRepurchaseTier                   = errors.New("active subscription tier can only be renewed")
	ErrSubscriptionActionRequiresActive = errors.New("subscription action requires an active subscription")
	ErrSubscriptionActionMismatch       = errors.New("subscription action does not match current subscription")
)
