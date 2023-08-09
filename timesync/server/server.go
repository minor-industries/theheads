package server

import (
	"context"
	gen "github.com/minor-industries/protobuf/gen/go/heads"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Handler struct {
	RTC bool
}

func (h *Handler) Time(ctx context.Context, in *gen.TimeIn) (*gen.TimeOut, error) {
	return &gen.TimeOut{
		T:      timestamppb.New(time.Now()),
		HasRtc: h.RTC,
	}, nil
}
