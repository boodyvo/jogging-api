package mongo

import (
	"time"

	"github.com/boodyvo/jogging-api/services/api/storage"
	"github.com/boodyvo/jogging-api/services/api/storage/mongo/filterparser"
	"github.com/google/uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	defaultTrackingPerRequest = 10
	defaultDuration           = 7 * 24 * time.Hour

	trackingCollection = "trackings"
)

func (d *database) SaveTracking(tracking *storage.Tracking) error {
	if err := d.session.DB(d.name).C(trackingCollection).Insert(tracking); err != nil {
		return err
	}

	return nil
}

func (d *database) UpdateTracking(tracking *storage.Tracking) error {
	if err := d.session.DB(d.name).C(trackingCollection).UpdateId(tracking.ID, tracking); err != nil {
		if err == mgo.ErrNotFound {
			return storage.ErrNotFound
		}
		return err
	}

	return nil
}

func (d *database) DeleteTracking(id uuid.UUID) error {
	if err := d.session.DB(d.name).C(trackingCollection).RemoveId(id); err != nil {
		if err == mgo.ErrNotFound {
			return nil
		}
		return err
	}

	return nil
}

func (d *database) GetTracking(id uuid.UUID) (*storage.Tracking, error) {
	var tracking storage.Tracking
	if err := d.session.DB(d.name).C(trackingCollection).FindId(id).One(&tracking); err != nil {
		if err == mgo.ErrNotFound {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	return &tracking, nil
}

func (d *database) ListTrackings(filter *storage.TrackingFilter) (*storage.ListTrackingsResponse, error) {
	var trackings []*storage.Tracking
	var err error
	query := bson.D{}

	if filter.Query != "" {
		query, err = filterparser.ParseTracking(filter.Query)
		if err != nil {
			return nil, err
		}
	}

	mongoQuery := d.session.DB(d.name).C(trackingCollection).Find(query)
	total, err := mongoQuery.Count()
	if err != nil {
		return nil, err
	}

	if filter.Cursor != "" {
		if !bson.IsObjectIdHex(filter.Cursor) {
			return nil, storage.ErrInvalidCursor
		}

		query = bson.D{{"$and", []bson.D{
			query,
			{{"cursor", bson.D{{"$gt", bson.ObjectIdHex(filter.Cursor)}}}},
		}}}
	}
	mongoQuery = d.session.DB(d.name).C(trackingCollection).Find(query)
	mongoQuery.Sort("cursor")

	limit := defaultTrackingPerRequest
	if filter.PerRequest != 0 {
		limit = int(filter.PerRequest)
	}
	mongoQuery.Limit(limit)

	if err := mongoQuery.All(&trackings); err != nil {
		return nil, err
	}

	return &storage.ListTrackingsResponse{
		Total:     int64(total),
		Trackings: trackings,
	}, nil
}

func (d *database) ListTrackingsForUser(filter *storage.TrackingFilter) (*storage.ListTrackingsResponse, error) {
	var trackings []*storage.Tracking
	var err error
	query := bson.D{}

	if filter.Query != "" {
		query, err = filterparser.ParseTracking(filter.Query)
		if err != nil {
			return nil, err
		}
	}

	query = bson.D{{"$and", []bson.D{
		query,
		{{"user_id", bson.D{{"$eq", filter.UserID}}}},
	}}}

	mongoQuery := d.session.DB(d.name).C(trackingCollection).Find(query)
	total, err := mongoQuery.Count()
	if err != nil {
		return nil, err
	}

	if filter.Cursor != "" {
		if !bson.IsObjectIdHex(filter.Cursor) {
			return nil, storage.ErrInvalidCursor
		}

		query = bson.D{{"$and", []bson.D{
			query,
			{{"cursor", bson.D{{"$gt", bson.ObjectIdHex(filter.Cursor)}}}},
		}}}
	}
	mongoQuery = d.session.DB(d.name).C(trackingCollection).Find(query)
	mongoQuery.Sort("cursor")

	limit := defaultTrackingPerRequest
	if filter.PerRequest != 0 {
		limit = int(filter.PerRequest)
	}
	mongoQuery.Limit(limit)

	if err := mongoQuery.All(&trackings); err != nil {
		return nil, err
	}

	return &storage.ListTrackingsResponse{
		Total:     int64(total),
		Trackings: trackings,
	}, nil
}

func (d *database) GetReport(filter *storage.ReportFilter) (*storage.Report, error) {
	start := filter.FromDate
	duration := defaultDuration
	if filter.Duration != time.Duration(0) {
		duration = filter.Duration
	}
	end := start.Add(duration)

	pipeline := []bson.M{
		{
			"$match": bson.M{"user_id": filter.UserID},
		},
		{
			"$match": bson.D{{"date", bson.D{{"$gte", start}}}},
		},
		{
			"$match": bson.D{{"date", bson.D{{"$lt", end}}}},
		},
		{
			"$group": bson.M{
				"_id":        "time",
				"time":       bson.M{"$sum": "$time"},
				"distance":   bson.M{"$sum": "$distance"},
				"time_count": bson.M{"$sum": 1},
			},
		},
	}

	result := make([]bson.M, 0)
	col := d.session.DB(d.name).C(trackingCollection)
	if err := col.Pipe(pipeline).All(&result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return &storage.Report{
			AverageSpeed: 0,
			Distance:     0,
		}, nil
	}

	resTimeInt64, ok := result[0]["time"].(int64)
	if !ok {
		return nil, ErrCannotCreateReport
	}
	resTime := time.Duration(resTimeInt64)
	resDistance, ok := result[0]["distance"].(float64)
	if !ok {
		return nil, ErrCannotCreateReport
	}
	averageSpeed := resDistance / resTime.Seconds()

	return &storage.Report{
		AverageSpeed: float32(averageSpeed),
		Distance:     float32(resDistance),
	}, nil
}
