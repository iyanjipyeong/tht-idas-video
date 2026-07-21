package entity

import "testing"

func TestCanAccessVideo(t *testing.T) {
	tests := []struct {
		name          string
		userTier      Tier
		videoCategory Tier
		want          bool
	}{
		{name: "gold accesses gold", userTier: TierGold, videoCategory: TierGold, want: true},
		{name: "gold accesses silver", userTier: TierGold, videoCategory: TierSilver, want: true},
		{name: "gold accesses bronze", userTier: TierGold, videoCategory: TierBronze, want: true},
		{name: "silver cannot access gold", userTier: TierSilver, videoCategory: TierGold, want: false},
		{name: "silver accesses silver", userTier: TierSilver, videoCategory: TierSilver, want: true},
		{name: "silver accesses bronze", userTier: TierSilver, videoCategory: TierBronze, want: true},
		{name: "bronze cannot access gold", userTier: TierBronze, videoCategory: TierGold, want: false},
		{name: "bronze cannot access silver", userTier: TierBronze, videoCategory: TierSilver, want: false},
		{name: "bronze accesses bronze", userTier: TierBronze, videoCategory: TierBronze, want: true},
		{name: "invalid user tier denied", userTier: TierUnknown, videoCategory: TierBronze, want: false},
		{name: "invalid video tier denied", userTier: TierGold, videoCategory: Tier(99), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanAccessVideo(tt.userTier, tt.videoCategory)
			if got != tt.want {
				t.Fatalf("CanAccessVideo() = %v, want %v", got, tt.want)
			}
		})
	}
}
