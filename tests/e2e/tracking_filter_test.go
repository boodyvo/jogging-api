// +build integration

package e2e

import (
	"testing"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"

	"github.com/boodyvo/jogging-api/tests/common"
	"github.com/boodyvo/jogging-api/tests/lib"
	"github.com/stretchr/testify/require"
)

const (
	testTrackings              = 25
	testQueryTrackingsSimple   = 5
	testQueryTrackingsMultiple = 15

	defaultTrackingPerReq = 10
)

func TestListOwnTracking(t *testing.T) {
	r := require.New(t)
	client := lib.NewClient(common.DefaultURL)

	user, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create user")

	trackings := make([]*lib.Tracking, 0, testTrackings)
	for i := 0; i < testTrackings; i++ {
		tracking, err := client.CreateRandomTracking(user)
		r.NoError(err, "cannot create tracking")

		trackings = append(trackings, tracking)
	}

	protoTrackings := make(map[string]*pb.Tracking, testTrackings)
	listTrackingResp, err := client.ListOwnTrackings(user, &pb.ListTrackingsRequest{})
	r.NoError(err, "cannot list trackings")
	r.Equal(int64(len(trackings)), listTrackingResp.Total, "incorrect total")
	r.Equal(defaultTrackingPerReq, len(listTrackingResp.Trackings), "incorrect total")
	for _, tracking := range listTrackingResp.Trackings {
		protoTrackings[tracking.Id] = tracking
	}
	cursor := listTrackingResp.Cursor

	listTrackingResp, err = client.ListOwnTrackings(user, &pb.ListTrackingsRequest{
		Cursor: cursor,
	})
	r.NoError(err, "cannot list trackings")
	r.Equal(int64(len(trackings)), listTrackingResp.Total, "incorrect total")
	r.Equal(defaultTrackingPerReq, len(listTrackingResp.Trackings), "incorrect total")
	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.False(ok, "same tracking for list request")
		protoTrackings[tracking.Id] = tracking
	}
	cursor = listTrackingResp.Cursor

	listTrackingResp, err = client.ListOwnTrackings(user, &pb.ListTrackingsRequest{
		PerReq: defaultTrackingPerReq + defaultTrackingPerReq/2,
	})
	r.NoError(err, "cannot list trackings")
	r.Equal(int64(len(trackings)), listTrackingResp.Total, "incorrect total")
	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.True(ok, "cannot find tracking")
	}

	listTrackingResp, err = client.ListOwnTrackings(user, &pb.ListTrackingsRequest{
		Cursor: cursor,
	})
	r.NoError(err, "cannot list trackings")
	r.Equal(int64(len(trackings)), listTrackingResp.Total, "incorrect total")
	r.Equal(testTrackings-2*defaultTrackingPerReq, len(listTrackingResp.Trackings), "incorrect total")
	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.False(ok, "same tracking for list request")
		protoTrackings[tracking.Id] = tracking
	}

	for _, tracking := range trackings {
		_, ok := protoTrackings[tracking.ID]
		r.True(ok, "cannot find tracking")
	}

	listTrackingResp, err = client.ListOwnTrackings(user, &pb.ListTrackingsRequest{
		PerReq: testTrackings,
	})
	r.NoError(err, "cannot list trackings")
	r.Equal(int64(len(trackings)), listTrackingResp.Total, "incorrect total")
	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.True(ok, "cannot find tracking")
	}
}

func TestListOwnTrackingSimpleQuery(t *testing.T) {
	r := require.New(t)
	client := lib.NewClient(common.DefaultURL)

	user, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create user")

	trackings := make([]*lib.Tracking, 0, testQueryTrackingsSimple)
	for i := 0; i < testQueryTrackingsSimple; i++ {
		tracking, err := client.CreateRandomTracking(user)
		r.NoError(err, "cannot create tracking")

		trackings = append(trackings, tracking)
	}

	protoTrackings := make(map[string]*pb.Tracking, testQueryTrackingsSimple)
	listTrackingResp, err := client.ListOwnTrackings(user, &pb.ListTrackingsRequest{
		Query: "(distance gt 500)",
	})
	r.NoError(err, "cannot list trackings")
	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.Less(float32(500), tracking.Distance, "distance should be more than 500")
		r.False(ok, "same tracking for list request")
		protoTrackings[tracking.Id] = tracking
	}

	listTrackingResp, err = client.ListOwnTrackings(user, &pb.ListTrackingsRequest{
		Query: "(distance lt 500.0001)",
	})
	r.NoError(err, "cannot list trackings")
	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.Less(tracking.Distance, float32(500.0001), "distance should be less than 500.0001")
		r.False(ok, "same tracking for list request")
		protoTrackings[tracking.Id] = tracking
	}

	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.True(ok, "cannot find tracking")
	}

	r.Equal(testQueryTrackingsSimple, len(protoTrackings), "not all trackings")
}

func TestListOwnTrackingQuery(t *testing.T) {
	r := require.New(t)
	client := lib.NewClient(common.DefaultURL)

	user, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create user")

	trackings := make([]*lib.Tracking, 0, testQueryTrackingsMultiple)
	for i := 0; i < testQueryTrackingsMultiple; i++ {
		tracking, err := client.CreateRandomTracking(user)
		r.NoError(err, "cannot create tracking")

		trackings = append(trackings, tracking)
	}

	protoTrackings := make(map[string]*pb.Tracking, testQueryTrackingsMultiple)

	listTrackingResp, err := client.ListOwnTrackings(user, &pb.ListTrackingsRequest{
		Query:  "(distance gt 500) and location.latitude lt 100.001",
		PerReq: testQueryTrackingsMultiple,
	})
	r.NoError(err, "cannot list trackings")
	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.Less(tracking.Location.Latitude, float64(100.0001), "latitude should be less than 100.001")
		r.Less(float32(500), tracking.Distance, "distance should be greater than 500")
		r.False(ok, "same tracking for list request")
		protoTrackings[tracking.Id] = tracking
	}

	listTrackingResp, err = client.ListOwnTrackings(user, &pb.ListTrackingsRequest{
		Query:  "(distance lt 500.0001) and location.latitude lt 100.001",
		PerReq: testQueryTrackingsMultiple,
	})
	r.NoError(err, "cannot list trackings")
	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.Less(tracking.Location.Latitude, float64(100.0001), "latitude should be less than 100.001")
		r.Less(tracking.Distance, float32(500.0001), "distance should be less than 500.0001")
		r.False(ok, "same tracking for list request")
		protoTrackings[tracking.Id] = tracking
	}

	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.True(ok, "cannot find tracking")
	}

	listTrackingResp, err = client.ListOwnTrackings(user, &pb.ListTrackingsRequest{
		Query:  "(distance gt 500) and location.latitude gt 100",
		PerReq: testQueryTrackingsMultiple,
	})
	r.NoError(err, "cannot list trackings")
	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.Less(float64(100), tracking.Location.Latitude, "latitude should be greater than 100")
		r.Less(float32(500), tracking.Distance, "distance should be greater than 500")
		r.False(ok, "same tracking for list request")
		protoTrackings[tracking.Id] = tracking
	}

	listTrackingResp, err = client.ListOwnTrackings(user, &pb.ListTrackingsRequest{
		Query:  "(distance lt 500.0001) and location.latitude gt 100",
		PerReq: testQueryTrackingsMultiple,
	})
	r.NoError(err, "cannot list trackings")
	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.Less(float64(100), tracking.Location.Latitude, "latitude should be greater than 100")
		r.Less(tracking.Distance, float32(500.0001), "distance should be less than 500.0001")
		r.False(ok, "same tracking for list request")
		protoTrackings[tracking.Id] = tracking
	}

	for _, tracking := range listTrackingResp.Trackings {
		_, ok := protoTrackings[tracking.Id]
		r.True(ok, "cannot find tracking")
	}
	r.Equal(testQueryTrackingsMultiple, len(protoTrackings), "not all trackings")
}
