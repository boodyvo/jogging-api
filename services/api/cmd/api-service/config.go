package main

import (
	"github.com/jessevdk/go-flags"
)

type Config struct {
	Port         int    `long:"port"`
	MongoUrl     string `long:"mongo"`
	DatabaseName string `long:"db_name"`
	PrivateKey   string `long:"private_key"`
	WeatherAppID string `json:"app_id"`
}

func parseConfig() (*Config, error) {
	config := &Config{
		Port:         9090,
		MongoUrl:     "mongodb://mongo:27017",
		DatabaseName: "jogging",
		PrivateKey: `-----BEGIN PRIVATE KEY-----
MIGkAgEBBDAzp+fBFNJ/8hwgny1dde/Go1ta6vjhXY/+FRrWALhiTNzvdJ6QXFIT
kWRWJLRwZhWgBwYFK4EEACKhZANiAAT0UBg3bj/axz6trLhSizkSng6T+0QlA1pq
zKzqpRWLVu38pLUGxkDOYr37D9RVotPua960GeLX+Kh/t8A9fO3fIz1NFj32IFSe
uN+j2rLcnHhrFrv05JXHDByimveEvAc=
-----END PRIVATE KEY-----`,
		WeatherAppID: "xSvXgS3B",
	}

	_, err := flags.Parse(config)

	return config, err
}
