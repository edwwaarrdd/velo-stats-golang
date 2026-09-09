package support_test

import (
	"testing"

	"velostats/internal/support"
)

func TestMoneyRoundsToTwoDecimals(t *testing.T) {
	cases := map[string]struct {
		value float64
		want  float64
	}{
		"rounds down":         {1.234, 1.23},
		"rounds up":           {1.236, 1.24},
		"keeps whole numbers": {10, 10.0},
		"rounds the double, not the typed decimal":      {(1599.5 / 1000) / (6.0 / 60), 15.99},
		"rounds 2.675 down":                             {2.675, 2.67},
		"rounds 1.585 down":                             {1.585, 1.58},
		"rounds an exact midpoint to the even digit":    {0.125, 0.12},
		"rounds an exact midpoint up to the even digit": {0.375, 0.38},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			if got := support.Money(testCase.value); got != testCase.want {
				t.Errorf("Money(%v) = %v, want %v", testCase.value, got, testCase.want)
			}
		})
	}
}

func TestMoneyPtrPassesAbsentValuesThrough(t *testing.T) {
	if support.MoneyPtr(nil) != nil {
		t.Error("MoneyPtr(nil) should stay absent")
	}
}
