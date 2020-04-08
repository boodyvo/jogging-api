package main

import (
	"github.com/jessevdk/go-flags"
)

type Config struct {
	Port int `long:"port"`
}

func parseConfig() (*Config, error) {
	config := &Config{
		Port: 8080,
	}

	_, err := flags.Parse(config)

	return config, err
}
