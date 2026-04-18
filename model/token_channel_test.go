package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenChannel_GetRemainingPercent(t *testing.T) {
	tests := []struct {
		name       string
		quotaLimit int64
		usedQuota  int64
		expected   int
	}{
		{"unlimited", 0, 0, 100},
		{"full", 100, 0, 100},
		{"half", 100, 50, 50},
		{"exhausted", 100, 100, 0},
		{"over", 100, 150, 0},
		{"zero_limit", 0, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &TokenChannel{
				QuotaLimit: tt.quotaLimit,
				UsedQuota:  tt.usedQuota,
			}
			result := tc.GetRemainingPercent()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenChannel_GetRemainingQuota(t *testing.T) {
	tests := []struct {
		name       string
		quotaLimit int64
		usedQuota  int64
		expected   int64
	}{
		{"unlimited", 0, 0, -1},
		{"full", 100, 0, 100},
		{"half", 100, 50, 50},
		{"exhausted", 100, 100, 0},
		{"over", 100, 150, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &TokenChannel{
				QuotaLimit: tt.quotaLimit,
				UsedQuota:  tt.usedQuota,
			}
			result := tc.GetRemainingQuota()
			assert.Equal(t, tt.expected, result)
		})
	}
}
