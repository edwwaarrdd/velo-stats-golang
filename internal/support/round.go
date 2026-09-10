package support

import "strconv"

const MoneyPrecision = 2

// The rounding is applied to the exact binary value of the float rather than to
// the decimal a human would have typed: 15.995 is really 15.99499999999999957,
// so it reports as 15.99 rather than 16.0. Values that really are an exact
// midpoint, such as 0.125, round to the even digit.
func Money(value float64) float64 {
	rounded, err := strconv.ParseFloat(strconv.FormatFloat(value, 'f', MoneyPrecision, 64), 64)
	if err != nil {
		return value
	}

	return rounded
}

func MoneyPtr(value *float64) *float64 {
	if value == nil {
		return nil
	}

	rounded := Money(*value)

	return &rounded
}
