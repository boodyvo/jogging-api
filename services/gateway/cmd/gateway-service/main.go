package main

import (
	"context"
	"fmt"
	"net/http"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	config, err := parseConfig()
	if err != nil {
		log.Fatal("cannot parse the config", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	marshaller := &runtime.HTTPBodyMarshaler{
		Marshaler: &runtime.JSONPb{EmitDefaults: true, OrigName: true},
	}
	grpcMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaller))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = pb.RegisterAPIServiceHandlerFromEndpoint(ctx, grpcMux, "api:9090", opts)
	if err != nil {
		log.Fatal("cannot start api", err)
	}

	router := mux.NewRouter()
	router.PathPrefix("/docs/").Handler(
		http.StripPrefix("/docs/", http.FileServer(http.Dir("/docs/"))),
	)
	router.PathPrefix("/").Handler(grpcMux)

	log.Infof("start listening gateway-service on port %d", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router))
}
