package storage

import (
	"time"

	"github.com/boodyvo/jogging-api/lib"

	"github.com/golang/protobuf/ptypes/duration"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"
	"github.com/google/uuid"
	"gopkg.in/mgo.v2/bson"
)

type Location struct {
	Longitude float64 `json:"longitude" bson:"longitude"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
}

type Tracking struct {
	ID       uuid.UUID     `json:"id" bson:"_id"`
	UserID   uuid.UUID     `json:"user_id" bson:"user_id"`
	Location Location      `json:"location" bson:"location"`
	Date     time.Time     `json:"date" bson:"date"`
	Time     time.Duration `json:"time" bson:"time"`
	// Distance represents in meters
	Distance float32       `json:"distance" bson:"distance"`
	Weather  *Weather      `json:"weather" bson:"weather"`
	Cursor   bson.ObjectId `json:"-" bson:"cursor"`
}

func NewTrackingFromProto(tracking *pb.CreateTrackingRequest) (*Tracking, error) {
	trackingDate, err := time.Parse(lib.DateFormat, tracking.Date)
	if err != nil {
		return nil, err
	}
	return &Tracking{
		Cursor: bson.NewObjectId(),
		ID:     uuid.New(),
		Location: Location{
			Longitude: tracking.Location.Longitude,
			Latitude:  tracking.Location.Latitude,
		},
		Date:     trackingDate,
		Time:     time.Duration(tracking.Time.Seconds * int64(time.Second)),
		Distance: tracking.Distance,
		Weather:  &Weather{},
	}, nil
}

func (t *Tracking) ToProto() *pb.Tracking {
	return &pb.Tracking{
		Id:       t.ID.String(),
		UserId:   t.UserID.String(),
		Date:     t.Date.Format(lib.DateFormat),
		Time:     &duration.Duration{Seconds: int64(t.Time.Seconds())},
		Distance: t.Distance,
		Location: &pb.Location{
			Longitude: t.Location.Longitude,
			Latitude:  t.Location.Latitude,
		},
		Weather: t.Weather.ToProto(),
	}
}

type TrackingFilter struct {
	UserID     uuid.UUID
	PerRequest int64
	Cursor     string
	Query      string
}

func TrackingFilterFromProtoForUser(tracking *pb.ListTrackingsRequest, user *User) (*TrackingFilter, error) {
	track, err := TrackingFilterFromProto(tracking)
	if err != nil {
		return nil, err
	}
	track.UserID = user.ID

	return track, nil
}

func TrackingFilterFromProto(tracking *pb.ListTrackingsRequest) (*TrackingFilter, error) {
	return &TrackingFilter{
		Cursor:     tracking.Cursor,
		PerRequest: tracking.PerReq,
		Query:      tracking.Query,
	}, nil
}

type ListTrackingsResponse struct {
	Total     int64
	Trackings []*Tracking
}

func ProtoFromListTrackingsResponse(response *ListTrackingsResponse) *pb.ListTrackingsResponse {
	cursor := ""
	trackings := make([]*pb.Tracking, 0, len(response.Trackings))
	for _, tracking := range response.Trackings {
		trackings = append(trackings, tracking.ToProto())
		if tracking.Cursor.Hex() > cursor {
			cursor = tracking.Cursor.Hex()
		}
	}
	return &pb.ListTrackingsResponse{
		Cursor:    cursor,
		Total:     response.Total,
		Trackings: trackings,
	}
}

type ReportFilter struct {
	UserID   uuid.UUID
	FromDate time.Time
	Duration time.Duration
}

func NewReportFilterFromProtoForUser(request *pb.ReportRequest, user *User) (*ReportFilter, error) {
	fromDate, err := time.Parse(lib.DateFormat, request.FromDate)
	if err != nil {
		return nil, err
	}
	dur := time.Duration(0)
	if request.Duration != nil {
		dur = time.Duration(request.Duration.Seconds * int64(time.Second))
	}

	return &ReportFilter{
		UserID:   user.ID,
		FromDate: fromDate.UTC(),
		Duration: dur,
	}, nil
}

type Report struct {
	AverageSpeed float32 `json:"average_speed" bson:"average_speed"`
	Distance     float32 `json:"distance" bson:"distance"`
}

func (r *Report) ToProto() *pb.ReportResponse {
	return &pb.ReportResponse{
		AverageSpeed: r.AverageSpeed,
		Distance:     r.Distance,
	}
}
