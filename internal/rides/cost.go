package rides

import (
	"context"
	"fmt"
	"time"

	"velostats/internal/support"
)

const (
	AnnualSubscriptionPriceEUR = 58.0
	DaysPerYear                = 365
	DayPassPriceEUR            = 5.0
	WeekPassPriceEUR           = 12.0
)

type Cost struct {
	TotalRides                   int64             `json:"total_rides"`
	FirstRideDate                *string           `json:"first_ride_date"`
	LastRideDate                 *string           `json:"last_ride_date"`
	DateRangeDays                support.NullInt   `json:"date_range_days"`
	SubscriptionPriceEUR         support.Float     `json:"subscription_price_eur"`
	ProratedSubscriptionPriceEUR support.NullFloat `json:"prorated_subscription_price_eur"`
	CostPerRideEUR               support.NullFloat `json:"cost_per_ride_eur"`
	DayPassEquivalentEUR         support.NullFloat `json:"day_pass_equivalent_eur"`
	WeekPassEquivalentEUR        support.NullFloat `json:"week_pass_equivalent_eur"`
	MoneySavedVsDayPassesEUR     support.NullFloat `json:"money_saved_vs_day_passes_eur"`
	MoneySavedVsWeekPassesEUR    support.NullFloat `json:"money_saved_vs_week_passes_eur"`
}

type CostCalculator struct {
	rides *Repository
}

func NewCostCalculator(rides *Repository) *CostCalculator {
	return &CostCalculator{rides: rides}
}

func (c *CostCalculator) Calculate(ctx context.Context) (Cost, error) {
	checkoutTimes, err := c.rides.CheckoutTimes(ctx)
	if err != nil {
		return Cost{}, err
	}

	totalRides := int64(len(checkoutTimes))

	if totalRides == 0 {
		return emptyCost(), nil
	}

	first, last := checkoutTimes[0], checkoutTimes[0]
	rideDays := map[string]struct{}{}
	rideWeeks := map[string]struct{}{}

	for _, checkoutTime := range checkoutTimes {
		checkoutTime = checkoutTime.UTC()

		if checkoutTime.Before(first) {
			first = checkoutTime
		}

		if checkoutTime.After(last) {
			last = checkoutTime
		}

		rideDays[checkoutTime.Format(support.APIDateFormat)] = struct{}{}

		isoYear, isoWeek := checkoutTime.ISOWeek()
		rideWeeks[fmt.Sprintf("%d-%d", isoYear, isoWeek)] = struct{}{}
	}

	firstRideDate := startOfDay(first)
	lastRideDate := startOfDay(last)
	dateRangeDays := int64(lastRideDate.Sub(firstRideDate).Hours()/24) + 1

	proratedSubscriptionPrice := support.Money(AnnualSubscriptionPriceEUR * float64(dateRangeDays) / DaysPerYear)
	costPerRide := support.Money(proratedSubscriptionPrice / float64(totalRides))

	dayPassEquivalent := support.Money(float64(len(rideDays)) * DayPassPriceEUR)
	weekPassEquivalent := support.Money(float64(len(rideWeeks)) * WeekPassPriceEUR)

	firstFormatted := support.APIDate(firstRideDate)
	lastFormatted := support.APIDate(lastRideDate)

	return Cost{
		TotalRides:                   totalRides,
		FirstRideDate:                &firstFormatted,
		LastRideDate:                 &lastFormatted,
		DateRangeDays:                support.NullInt{Int: dateRangeDays, Valid: true},
		SubscriptionPriceEUR:         support.Float(AnnualSubscriptionPriceEUR),
		ProratedSubscriptionPriceEUR: support.FloatValue(proratedSubscriptionPrice),
		CostPerRideEUR:               support.FloatValue(costPerRide),
		DayPassEquivalentEUR:         support.FloatValue(dayPassEquivalent),
		WeekPassEquivalentEUR:        support.FloatValue(weekPassEquivalent),
		MoneySavedVsDayPassesEUR:     support.FloatValue(support.Money(dayPassEquivalent - proratedSubscriptionPrice)),
		MoneySavedVsWeekPassesEUR:    support.FloatValue(support.Money(weekPassEquivalent - proratedSubscriptionPrice)),
	}, nil
}

func emptyCost() Cost {
	return Cost{
		SubscriptionPriceEUR: support.Float(AnnualSubscriptionPriceEUR),
	}
}

func startOfDay(value time.Time) time.Time {
	value = value.UTC()

	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
