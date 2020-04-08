package main

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"

	"github.com/boodyvo/jogging-api/services/api/client"

	"github.com/urfave/cli/v2"
)

var userCommand = &cli.Command{
	Name:  "users",
	Usage: "Operation with users",
	Subcommands: []*cli.Command{
		{
			Name:  "createadmin",
			Usage: "Create admin user",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "email",
					Value: "",
					Usage: "email of admin user",
				},
				&cli.StringFlag{
					Name:  "password",
					Value: "",
					Usage: "password of admin user",
				},
			},
			Action: createAdminUser,
		},
	},
}

func createAdminUser(ctx *cli.Context) error {
	rpcaddr := ctx.String("rpcaddr")

	c, err := client.New(context.Background(), rpcaddr)
	if err != nil {
		return fmt.Errorf("unable to connect to api: %v", err)
	}

	email := ctx.String("email")
	password := ctx.String("password")

	resp, err := c.CreateAdmin(context.Background(), &pb.CreateAdminRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("cannot create admin user: %v", err)
	}

	printRespJSON(resp)
	return nil
}

func printRespJSON(resp proto.Message) {
	jsonMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "    ",
	}

	jsonStr, err := jsonMarshaler.MarshalToString(resp)
	if err != nil {
		fmt.Println("unable to decode response: ", err)
		return
	}

	fmt.Println(jsonStr)
}
