package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"

	"github.com/oklog/run"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"
	"github.com/boodyvo/jogging-api/services/api"
	"github.com/boodyvo/jogging-api/services/api/auth"
	"github.com/boodyvo/jogging-api/services/api/storage/mongo"
	"github.com/boodyvo/jogging-api/services/api/weather"
)

func main() {
	var group run.Group

	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})

	config, err := parseConfig()
	if err != nil {
		logger.Fatal("cannot parse the config", err)
	}

	store, err := mongo.New(config.MongoUrl, config.DatabaseName)
	if err != nil {
		logger.Fatal("cannot create a storage", err)
	}

	// TODO(boodyvo): Read from env/file, not from flag
	block, _ := pem.Decode([]byte(config.PrivateKey))
	x509Encoded := block.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		logger.Fatal("cannot parse private key", err)
	}

	authServer := auth.New(privateKey, store, logger)
	weatherServer := weather.NewService(config.WeatherAppID, logger)

	server := api.New(store, authServer, weatherServer, logger)

	s := grpc.NewServer()
	pb.RegisterAPIServiceServer(s, server)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	// TODO(boodyvo): Separate into another service with message broker
	group.Add(func() error {
		err := server.Start()
		logger.Infof("finish server: %v", err)

		return err
	}, func(err error) {})

	logger.Infof("start listening api-service on port %d", config.Port)
	group.Add(func() error {
		err := s.Serve(lis)
		logger.Infof("stop serving service: %v", err)

		return err
	}, func(err error) {})

	logger.Infof("api-service terminated: %v", group.Run())
}
