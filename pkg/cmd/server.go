package cmd

import (
	"log"

	"github.com/frauniki/echo-server/pkg/server"
	"github.com/spf13/cobra"
)

var (
	host           = "127.0.0.1"
	grpcListenPort = 8081
	httpListenPort = 8080
)

var serverCmd = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := server.Run(host, grpcListenPort, httpListenPort); err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	serverCmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "Listen Host")
	serverCmd.Flags().IntVarP(&grpcListenPort, "grpc-listen-port", "g", 8081, "gRPC Listen Port")
	serverCmd.Flags().IntVarP(&httpListenPort, "http-listen-port", "t", 8080, "HTTP Listen Port")
}
