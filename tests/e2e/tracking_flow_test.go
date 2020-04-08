// +build integration

package e2e

import (
	"testing"
	"time"

	lib2 "github.com/boodyvo/jogging-api/lib"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"

	"github.com/boodyvo/jogging-api/tests/common"

	"github.com/boodyvo/jogging-api/tests/lib"
	"github.com/stretchr/testify/require"
)

func TestTrackingFlow(t *testing.T) {
	r := require.New(t)
	client := lib.NewClient(common.DefaultURL)

	user, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create user")

	createTrackingRequest := &lib.CreateTrackingRequest{
		Location: lib.Location{
			Longitude: 20.2181231,
			Latitude:  50.6243121,
		},
		Time:     time.Hour.String(),
		Distance: 11.51,
		Date:     time.Now().AddDate(0, 0, -5),
	}
	createResp, err := client.CreateTracking(user, createTrackingRequest)
	r.NoError(err, "cannot create trackingResp")

	trackingResp, err := client.GetTracking(user, &pb.GetTrackingRequest{
		Id: createResp.Id,
	})
	r.NoError(err, "cannot get own trackingResp")
	r.Equal(createResp.Id, trackingResp.Tracking.Id)
	r.Equal(user.ID, trackingResp.Tracking.UserId)
	r.Equal(createTrackingRequest.Date.Format(lib2.DateFormat), trackingResp.Tracking.Date)
	r.Equal(
		createTrackingRequest.Time,
		time.Duration(trackingResp.Tracking.Time.Seconds*int64(time.Second)).String(),
	)
	r.Equal(createTrackingRequest.Location.Longitude, trackingResp.Tracking.Location.Longitude)
	r.Equal(createTrackingRequest.Location.Latitude, trackingResp.Tracking.Location.Latitude)

	userSecond, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create user")

	_, err = client.GetTracking(userSecond, &pb.GetTrackingRequest{
		Id: createResp.Id,
	})
	r.Error(err, "can get trackingResp of another user")
}
