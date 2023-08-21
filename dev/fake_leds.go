package main

import (
	"context"
	"github.com/minor-industries/protobuf/gen/go/heads"
	"github.com/minor-industries/theheads/common/standard_server"
	"github.com/minor-industries/theheads/common/util"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type fakeleds struct {
	fakeledsHandler
}

type fakeledsHandler struct {
}

func (l fakeledsHandler) Run(ctx context.Context, in *heads.RunIn) (*heads.Empty, error) {
	switch in.Name {
	case "highred", "rainbow":
		return &heads.Empty{}, nil
	default:
		return nil, errors.New("unknown animation")
	}
}

func (l fakeledsHandler) Events(empty *heads.Empty, server heads.Leds_EventsServer) error {
	//TODO implement me
	panic("implement me")
}

func (l fakeledsHandler) SetScale(ctx context.Context, in *heads.SetScaleIn) (*heads.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (l fakeledsHandler) Ping(ctx context.Context, empty *heads.Empty) (*heads.Empty, error) {
	return &heads.Empty{}, nil
}

func (l *fakeleds) Run(port int) {
	logger, _ := util.NewLogger(false)

	server, err := standard_server.NewServer(&standard_server.Config{
		Logger: logger,
		Port:   port,
		GrpcSetup: func(grpcServer *grpc.Server) error {
			heads.RegisterPingServer(grpcServer, l)
			heads.RegisterLedsServer(grpcServer, l.fakeledsHandler)
			return nil
		},
	})

	if err != nil {
		panic(err)
	}

	go server.Run()
}
