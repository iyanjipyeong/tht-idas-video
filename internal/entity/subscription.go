package entity

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type SubscriptionStatus int

const (
	SubscriptionStatusUnknown SubscriptionStatus = iota
	SubscriptionStatusActive
	SubscriptionStatusInactive
	SubscriptionStatusExpired
)

type TierSnapshot struct {
	TierID       UUID    `json:"tierId"`
	TierCode     Tier    `json:"tierCode"`
	TierName     string  `json:"tierName"`
	TierLevel    int     `json:"tierLevel"`
	TierPrice    float64 `json:"tierPrice"`
	TierCurrency string  `json:"tierCurrency"`
}

var subscriptionStatusNames = map[SubscriptionStatus]string{
	SubscriptionStatusActive:   SubscriptionStatusNameActive,
	SubscriptionStatusInactive: SubscriptionStatusNameInactive,
	SubscriptionStatusExpired:  SubscriptionStatusNameExpired,
}

var subscriptionStatusValues = map[string]SubscriptionStatus{
	SubscriptionStatusNameActive:   SubscriptionStatusActive,
	SubscriptionStatusNameInactive: SubscriptionStatusInactive,
	SubscriptionStatusNameExpired:  SubscriptionStatusExpired,
}

func (status SubscriptionStatus) IsValid() bool {
	_, ok := subscriptionStatusNames[status]
	return ok
}

func (status SubscriptionStatus) String() string {
	if name, ok := subscriptionStatusNames[status]; ok {
		return name
	}
	return SubscriptionStatusNameUnknown
}

func ParseSubscriptionStatus(value string) (SubscriptionStatus, error) {
	status, ok := subscriptionStatusValues[strings.ToLower(strings.TrimSpace(value))]
	if !ok {
		return SubscriptionStatusUnknown, fmt.Errorf("invalid subscription status %q", value)
	}
	return status, nil
}

func (status SubscriptionStatus) MarshalJSON() ([]byte, error) {
	if !status.IsValid() {
		return json.Marshal(EmptyString)
	}
	return json.Marshal(status.String())
}

func (status *SubscriptionStatus) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := ParseSubscriptionStatus(value)
	if err != nil {
		return err
	}

	*status = parsed
	return nil
}

type Subscription struct {
	ID           UUID
	UserID       UUID
	TierID       UUID
	TierCode     Tier
	TierSnapshot TierSnapshot
	Status       SubscriptionStatus
	StartDate    time.Time
	EndDate      time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (subscription Subscription) IsActive(now time.Time) bool {
	return subscription.Status == SubscriptionStatusActive && subscription.EndDate.After(now)
}
