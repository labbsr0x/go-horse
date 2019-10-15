package cmd

import (
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters"
	filterConfig "gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/config-filter"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web"
	webConfig "gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config-web"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the HTTP REST APIs server",
	RunE: func(cmd *cobra.Command, args []string) error {

		filterBuilder := new(filterConfig.FilterBuilder).InitFromViper(viper.GetViper())
		filter := new(filters.Filter).InitFromFilterBuilder(filterBuilder)
		filter.ListAPIs.Init()

		webBuilder := new(webConfig.WebBuilder).InitFromViper(viper.GetViper(), filter)
		server := new(web.Server).InitFromWebBuilder(webBuilder)

		return server.Run()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	webConfig.AddFlags(serveCmd.Flags())
	filterConfig.AddFlags(serveCmd.Flags())

	err := viper.GetViper().BindPFlags(serveCmd.Flags())
	if err != nil {
		panic(err)
	}
}
