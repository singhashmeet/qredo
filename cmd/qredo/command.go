package qredo

import (
	"github.com/singhashmeet/temp/api"
	"github.com/singhashmeet/temp/config"

	"github.com/spf13/cobra"
)

var flags = config.Config{}

var DaemonCommand = &cobra.Command{
	Use:          "daemon",
	Short:        "Starts the daemon",
	RunE:         run,
	SilenceUsage: true,
}

func init() {
	DaemonCommand.PersistentFlags().StringVar(&flags.Host, "host", "", "If set starts the HTTP listener on the given TCP address")
	DaemonCommand.PersistentFlags().StringVar(&flags.Port, "port", "8080", "If set starts the HTTP listener on the given TCP port")
}

func run(cmd *cobra.Command, args []string) (err error) {
	// inilize the server
	server := api.NewServer(flags.Host, flags.Port)
	// start the server
	server.ListenAndServe()
	return nil
}
