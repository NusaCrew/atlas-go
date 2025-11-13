package webserver

import (
	"context"
	"errors"

	"github.com/NusaCrew/atlas-go/log"

	"github.com/spf13/cobra"
)

type WebServer interface {
	Run(ctx context.Context, errorChannel chan error)
	GetName() string
	Stop()
}

func RunServersCommand(ctx context.Context, servers ...WebServer) *cobra.Command {
	return &cobra.Command{
		Use:   "web",
		Short: "start web servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(servers) == 0 {
				return errors.New("cannot run servers: no servers provided")
			}

			var err error
			errorChannel := make(chan error, len(servers))

			for _, server := range servers {
				log.Info("starting: %s server...", server.GetName())
				server.Run(ctx, errorChannel)
			}

			select {
			case err = <-errorChannel:
				log.Error("server ended with err: %s", err)
			case <-ctx.Done():
				log.Info("context cancelled")
			}

			for _, server := range servers {
				log.Info("stopping server: %s", server.GetName())
				server.Stop()
			}

			return err
		},
	}
}
