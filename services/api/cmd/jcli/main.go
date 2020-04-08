package main

import (
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()

	app.Name = "jcli"
	app.Usage = "Command line interface for Jogging application"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Commands = []*cli.Command{
		userCommand,
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "rpcaddr",
			Value: "localhost:9090",
			Usage: "host:port of grpc api to connect",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("cannot run app: %v", err)
	}
}
