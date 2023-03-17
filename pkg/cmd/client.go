package cmd

import (
	"log"
	"time"

	"github.com/frauniki/echo-server/pkg/client"
	"github.com/spf13/cobra"
)

var (
	serverAddress string
	loop          bool
	tickSecond    int
	stream        bool
)

var clientCmd = &cobra.Command{
	Use: "client",
	Run: func(cmd *cobra.Command, args []string) {
		if tickSecond <= 0 {
			log.Fatal("tick option must be positive")
		}

		if err := client.Run(serverAddress, stream, loop, time.Duration(tickSecond)*time.Second); err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	clientCmd.Flags().StringVarP(&serverAddress, "server", "s", "127.0.0.1:8081", "Server Address")
	clientCmd.Flags().BoolVarP(&loop, "loop", "l", false, "")
	clientCmd.Flags().IntVarP(&tickSecond, "tick-second", "t", 1, "")
	clientCmd.Flags().BoolVarP(&stream, "stream", "S", false, "")
}
