package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/boodyvo/jogging-api/lib"

	log "github.com/sirupsen/logrus"

	store "github.com/boodyvo/jogging-api/services/api/storage"
)

type Service interface {
	GetWeather(time time.Time, location store.Location) (*store.Weather, error)
}

type OpenWeatherMapService struct {
	apiKey string
	logger *log.Logger
	client *http.Client
}

func NewService(apiKey string, logger *log.Logger) Service {
	return &OpenWeatherMapService{
		apiKey: apiKey,
		logger: logger,
		client: &http.Client{},
	}
}

func (s *OpenWeatherMapService) GetNearestStationID(location store.Location) (string, error) {
	URL := "https://api.meteostat.net/v1/stations/nearby?lat=%f&lon=%f&limit=1&key=%s"
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			URL,
			location.Latitude,
			location.Longitude,
			s.apiKey,
		),
		nil,
	)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result StationResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", ErrCannotGetWeather
	}
	return result.Data[0].ID, nil
}

// TODO(boodyvo): Get historical data, not current, because it's need paid subscription https://openweathermap.org/price
func (s *OpenWeatherMapService) GetWeather(time time.Time, location store.Location) (*store.Weather, error) {
	// baseURL https://api.meteostat.net/v1/history/daily?station=48694&start=2020-03-23&end=2020-03-24&key=xSvXgS3B
	station, err := s.GetNearestStationID(location)
	if err != nil {
		return nil, err
	}

	URL := "https://api.meteostat.net/v1/history/daily?station=%s&start=%s&end=%s&key=%s"
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			URL,
			station,
			time.UTC().Format(lib.DateFormat),
			time.UTC().Format(lib.DateFormat),
			s.apiKey,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Response

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, ErrCannotGetWeather
	}
	return &result.Data[0], nil
}
