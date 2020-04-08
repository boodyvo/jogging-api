package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/boodyvo/jogging-api/tests/common"
	"github.com/boodyvo/jogging-api/tests/lib"
)

const (
	trackingReportCount    = 3
	trackingReportDistance = 10.1
	trackingReportTime     = time.Hour

	delta = 0.001
)

func min(x, y float32) float32 {
	if x < y {
		return x
	}

	return y
}

func max(x, y float32) float32 {
	if x > y {
		return x
	}

	return y
}

func TestTrackingReportDefault(t *testing.T) {
	r := require.New(t)
	client := lib.NewClient(common.DefaultURL)

	user, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create user")

	trackingDateNow := time.Now().AddDate(0, -6, 0).UTC()

	for days := -3; days < 10; days++ {
		trackingDayInDays := trackingDateNow.AddDate(0, 0, days)
		createTrackingRequest := &lib.CreateTrackingRequest{
			Location: lib.CreateLocation(),
			Time:     trackingReportTime.String(),
			Distance: trackingReportDistance,
			Date:     trackingDayInDays,
		}
		for i := 0; i < trackingReportCount; i++ {
			_, err = client.CreateTracking(user, createTrackingRequest)
			r.NoError(err, "cannot create tracking")
		}

		reportResp, err := client.Report(user, &lib.ReportRequest{
			FromDate: trackingDateNow,
		})
		r.NoError(err, "cannot create report")
		multiplier := max(min(float32(days+1), 7), 0)

		distance := multiplier * float32(trackingReportDistance*trackingReportCount)
		trackingTime := multiplier * float32(trackingReportCount*trackingReportTime.Seconds())
		averageSpeed := float32(0)
		if trackingTime > delta {
			averageSpeed = distance / trackingTime
		}
		r.InDelta(distance, reportResp.Distance, delta, fmt.Sprintf("incorrect distance for day %d", days))
		r.InDelta(averageSpeed, reportResp.AverageSpeed, delta, fmt.Sprintf("incorrect speed for day %d", days))
	}
}

func TestTrackingReportCustomDuration(t *testing.T) {
	r := require.New(t)
	client := lib.NewClient(common.DefaultURL)

	user, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create user")

	trackingDateNow := time.Now().AddDate(0, -6, 0).UTC()
	numberOfDays := int64(14)

	for days := -3; days < 20; days++ {
		trackingDayInDays := trackingDateNow.AddDate(0, 0, days)
		createTrackingRequest := &lib.CreateTrackingRequest{
			Location: lib.CreateLocation(),
			Time:     trackingReportTime.String(),
			Distance: trackingReportDistance,
			Date:     trackingDayInDays,
		}
		for i := 0; i < trackingReportCount; i++ {
			_, err = client.CreateTracking(user, createTrackingRequest)
			r.NoError(err, "cannot create tracking")
		}

		reportResp, err := client.Report(user, &lib.ReportRequest{
			FromDate: trackingDateNow,
			// 2 weeks duration
			Duration: time.Duration(numberOfDays * 24 * int64(time.Hour)).String(),
		})
		r.NoError(err, "cannot create report")
		multiplier := max(min(float32(days+1), float32(numberOfDays)), 0)

		distance := multiplier * float32(trackingReportDistance*trackingReportCount)
		trackingTime := multiplier * float32(trackingReportCount*trackingReportTime.Seconds())
		averageSpeed := float32(0)
		if trackingTime > delta {
			averageSpeed = distance / trackingTime
		}
		r.InDelta(distance, reportResp.Distance, delta, fmt.Sprintf("incorrect distance for day %d", days))
		r.InDelta(averageSpeed, reportResp.AverageSpeed, delta, fmt.Sprintf("incorrect speed for day %d", days))
	}
}
