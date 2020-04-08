package weather

import store "github.com/boodyvo/jogging-api/services/api/storage"

type Station struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type StationResponse struct {
	Data []Station `json:"data"`
}

type Response struct {
	Data []store.Weather `json:"data"`
}
