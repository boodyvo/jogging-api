package lib

import (
	"time"
)

type User struct {
	ID           string
	Email        string
	Password     string
	AccessToken  string
	RefreshToken string
}

type Location struct {
	Longitude float64 `json:"longitude" bson:"longitude"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
}

type CreateTrackingRequest struct {
	Location Location  `json:"location" bson:"location"`
	Date     time.Time `json:"date" bson:"date"`
	Time     string    `json:"time" bson:"time"`
	Distance float32   `json:"distance" bson:"distance"`
}
type createTrackingRequestSerialized struct {
	Location Location `json:"location" bson:"location"`
	Date     string   `json:"date" bson:"date"`
	Time     string   `json:"time" bson:"time"`
	Distance float32  `json:"distance" bson:"distance"`
}

type Tracking struct {
	ID       string    `json:"id" bson:"_id"`
	UserID   string    `json:"user_id" bson:"user_id"`
	Location Location  `json:"location" bson:"location"`
	Date     time.Time `json:"date" bson:"date"`
	Time     string    `json:"time" bson:"time"`
	Distance float32   `json:"distance" bson:"distance"`
}

type ReportRequest struct {
	FromDate time.Time `json:"from_date" bson:"from_date"`
	Duration string    `json:"duration" bson:"duration"`
}
