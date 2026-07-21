package entity

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Tier int

const (
	TierUnknown Tier = iota
	TierBronze
	TierSilver
	TierGold
)

type TierDetail struct {
	ID          UUID    `json:"id"`
	Code        Tier    `json:"code"`
	Name        string  `json:"name"`
	Level       int     `json:"level"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

func (tierDetail TierDetail) ToSnapshot() TierSnapshot {
	return TierSnapshot{
		TierID:       tierDetail.ID,
		TierCode:     tierDetail.Code,
		TierName:     tierDetail.Name,
		TierLevel:    tierDetail.Level,
		TierPrice:    tierDetail.Price,
		TierCurrency: tierDetail.Currency,
	}
}

var tierNames = map[Tier]string{
	TierBronze: "Bronze",
	TierSilver: "Silver",
	TierGold:   "Gold",
}

var tierLevels = map[Tier]int{
	TierBronze: 1,
	TierSilver: 2,
	TierGold:   3,
}

var tierPrices = map[Tier]float64{
	TierBronze: 50000,
	TierSilver: 100000,
	TierGold:   150000,
}

var tierDescriptions = map[Tier]string{
	TierBronze: "Can access Bronze videos only",
	TierSilver: "Can access Silver and Bronze videos",
	TierGold:   "Can access Gold, Silver, and Bronze videos",
}

var tierValues = map[string]Tier{
	"bronze": TierBronze,
	"silver": TierSilver,
	"gold":   TierGold,
}

func (tier Tier) IsValid() bool {
	_, ok := tierNames[tier]
	return ok
}

func (tier Tier) String() string {
	if name, ok := tierNames[tier]; ok {
		return name
	}
	return "Unknown"
}

func (tier Tier) Level() int {
	return tierLevels[tier]
}

func (tier Tier) Price() float64 {
	return tierPrices[tier]
}

func (tier Tier) Currency() string {
	return "IDR"
}

func (tier Tier) Detail() TierDetail {
	return TierDetail{
		Code:        tier,
		Name:        tier.String(),
		Level:       tier.Level(),
		Price:       tier.Price(),
		Currency:    tier.Currency(),
		Description: tierDescriptions[tier],
	}
}

func AvailableTiers() []TierDetail {
	return []TierDetail{TierBronze.Detail(), TierSilver.Detail(), TierGold.Detail()}
}

func ParseTier(value string) (Tier, error) {
	tier, ok := tierValues[strings.ToLower(strings.TrimSpace(value))]
	if !ok {
		return TierUnknown, fmt.Errorf("invalid tier %q", value)
	}
	return tier, nil
}

func (tier Tier) MarshalJSON() ([]byte, error) {
	if !tier.IsValid() {
		return json.Marshal("")
	}
	return json.Marshal(tier.String())
}

func (tier *Tier) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := ParseTier(value)
	if err != nil {
		return err
	}

	*tier = parsed
	return nil
}

func CanAccessVideo(userTier Tier, videoCategory Tier) bool {
	if !userTier.IsValid() || !videoCategory.IsValid() {
		return false
	}

	return userTier >= videoCategory
}
