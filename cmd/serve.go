package server

import (
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the HTTP REST APIs server",
	RunE: func(cmd *cobra.Command, args []string) error {

		builder := new(config.WebBuilder).Init(viper.GetViper(), mailChannel)

		server := new(web.Server).InitFromWebBuilder(builder)

		err = server.Run()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	//config.AddFlags(serveCmd.Flags())

	err := viper.GetViper().BindPFlags(serveCmd.Flags())
	if err != nil {
		panic(err)
	}
}
