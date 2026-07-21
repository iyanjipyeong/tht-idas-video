package entity

import "testing"

func TestIsUUID(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "valid v4 lowercase", value: "11111111-1111-4111-8111-111111111111", want: true},
		{name: "valid v4 uppercase", value: "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", want: true},
		{name: "missing hyphen", value: "11111111111141118111111111111111", want: false},
		{name: "invalid version", value: "11111111-1111-6111-8111-111111111111", want: false},
		{name: "invalid variant", value: "11111111-1111-4111-1111-111111111111", want: false},
		{name: "empty", value: "", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsUUID(test.value); got != test.want {
				t.Fatalf("IsUUID(%q) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}
