package httpapi_test

import (
	"testing"

	"velostats/internal/rides"
	"velostats/internal/testsupport"
)

func TestCostReportsNoBreakdownWhenThereAreNoRides(t *testing.T) {
	_, body := get(t, testsupport.NewDatabase(t), "/rides/cost")

	assertJSON(t, body, `{
		"total_rides":0,
		"first_ride_date":null,
		"last_ride_date":null,
		"date_range_days":null,
		"subscription_price_eur":58.0,
		"prorated_subscription_price_eur":null,
		"cost_per_ride_eur":null,
		"day_pass_equivalent_eur":null,
		"week_pass_equivalent_eur":null,
		"money_saved_vs_day_passes_eur":null,
		"money_saved_vs_week_passes_eur":null
	}`)
}

func TestCostProratesTheSubscriptionOverTheRideDateRange(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Ride(t, db, rides.Ride{CheckoutTime: testsupport.Time(t, "2026-01-01 08:00:00")})
	testsupport.Ride(t, db, rides.Ride{CheckoutTime: testsupport.Time(t, "2026-01-01 12:00:00")})
	testsupport.Ride(t, db, rides.Ride{CheckoutTime: testsupport.Time(t, "2026-01-10 08:00:00")})

	_, body := get(t, db, "/rides/cost")

	assertJSON(t, body, `{
		"total_rides":3,
		"first_ride_date":"2026-01-01",
		"last_ride_date":"2026-01-10",
		"date_range_days":10,
		"subscription_price_eur":58.0,
		"prorated_subscription_price_eur":1.59,
		"cost_per_ride_eur":0.53,
		"day_pass_equivalent_eur":10.0,
		"week_pass_equivalent_eur":24.0,
		"money_saved_vs_day_passes_eur":8.41,
		"money_saved_vs_week_passes_eur":22.41
	}`)
}

func TestCostCountsASingleRideAsAOneDayRange(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Ride(t, db, rides.Ride{CheckoutTime: testsupport.Time(t, "2026-05-04 09:30:00")})

	_, body := get(t, db, "/rides/cost")
	cost := summary(t, body)

	expected := map[string]any{
		"date_range_days":                 1.0,
		"first_ride_date":                 "2026-05-04",
		"last_ride_date":                  "2026-05-04",
		"prorated_subscription_price_eur": 0.16,
		"cost_per_ride_eur":               0.16,
		"day_pass_equivalent_eur":         5.0,
		"week_pass_equivalent_eur":        12.0,
	}

	for field, want := range expected {
		if cost[field] != want {
			t.Errorf("%s = %v, want %v", field, cost[field], want)
		}
	}
}

func TestCostCountsRidesInTheSameISOWeekOnlyOnce(t *testing.T) {
	db := testsupport.NewDatabase(t)

	// Monday and Sunday of the same ISO week, so two ride days but one week.
	testsupport.Ride(t, db, rides.Ride{CheckoutTime: testsupport.Time(t, "2026-03-02 08:00:00")})
	testsupport.Ride(t, db, rides.Ride{CheckoutTime: testsupport.Time(t, "2026-03-08 08:00:00")})

	_, body := get(t, db, "/rides/cost")
	cost := summary(t, body)

	if cost["day_pass_equivalent_eur"] != 10.0 {
		t.Errorf("day_pass_equivalent_eur = %v, want 10", cost["day_pass_equivalent_eur"])
	}

	if cost["week_pass_equivalent_eur"] != 12.0 {
		t.Errorf("week_pass_equivalent_eur = %v, want 12", cost["week_pass_equivalent_eur"])
	}
}

func TestCostCountsRideDaysFromTheCheckoutTimeInUTC(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Ride(t, db, rides.Ride{CheckoutTime: testsupport.Time(t, "2026-06-01 23:30:00")})
	testsupport.Ride(t, db, rides.Ride{CheckoutTime: testsupport.Time(t, "2026-06-02 00:30:00")})

	_, body := get(t, db, "/rides/cost")
	cost := summary(t, body)

	if cost["first_ride_date"] != "2026-06-01" || cost["last_ride_date"] != "2026-06-02" || cost["date_range_days"] != 2.0 {
		t.Errorf("range = %v .. %v over %v days", cost["first_ride_date"], cost["last_ride_date"], cost["date_range_days"])
	}
}
