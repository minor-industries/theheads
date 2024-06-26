package heads_cli

import (
	"context"
	"fmt"
	"github.com/hashicorp/serf/client"
	"github.com/minor-industries/protobuf/gen/go/heads"
	"github.com/minor-industries/theheads/heads-cli/lib"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type volumeCmd struct {
	Match string `long:"match" description:"host pattern to match" default:"^head"`
	Vol   int    `long:"vol" description:"volume db to use" required:"true"`
}

func (opt *volumeCmd) Execute(args []string) error {
	return lib.ConnectAll(opt.Match, 8080, func(ctx context.Context, m *client.Member, conn *grpc.ClientConn) error {
		fmt.Println("setting volume for", m.Name)

		_, err := heads.NewVoicesClient(conn).SetVolume(ctx, &heads.SetVolumeIn{
			VolDb: int32(opt.Vol),
		})

		return errors.Wrap(err, "set volume")
	})
}
