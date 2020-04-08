package storage

import (
	pb "github.com/boodyvo/jogging-api/proto/pb/api"
)

type Weather struct {
	Temperature    float32 `json:"temperature" bson:"temperature"`
	TemperatureMin float32 `json:"temperature_min" bson:"temperature_min"`
	TemperatureMax float32 `json:"temperature_max" bson:"temperature_max"`
	Snowdepth      float32 `json:"snowdepth" bson:"snowdepth"`
	Winddirection  float32 `json:"winddirection" bson:"winddirection"`
	Windspeed      float32 `json:"windspeed" bson:"windspeed"`
	Pressure       float32 `json:"pressure" bson:"pressure"`
}

func (w *Weather) ToProto() *pb.Weather {
	return &pb.Weather{
		Temperature:    w.Temperature,
		TemperatureMin: w.TemperatureMin,
		TemperatureMax: w.TemperatureMax,
		Snowdepth:      w.Snowdepth,
		Winddirection:  w.Winddirection,
		Windspeed:      w.Windspeed,
		Pressure:       w.Pressure,
	}
}

type WeatherList struct {
	Cnt  int64     `json:"cnt"`
	List []Weather `json:"list"`
}
