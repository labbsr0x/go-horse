package cmd

import (
	"github.com/labbsr0x/go-horse/filters"
	filterConfig "github.com/labbsr0x/go-horse/filters/config-filter"
	"github.com/labbsr0x/go-horse/web"
	webConfig "github.com/labbsr0x/go-horse/web/config-web"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the HTTP REST APIs server",
	Example: ` 
./go-horse serve \
  --docker-api-version 1.39 \
  --docker-sock-url unix:///var/run/docker.sock \
  --target-host-name http://go-horse \
  --log-level info \
  --js-filters-path /app/go-horse/filters \
  --go-plugins-path /app/go-horse/plugins \
  --shutdown-time 5 \

All command line options can be provided via environment variables by adding the prefix "GOHORSE_" 
and converting their names to upper case and replacing punctuation and hyphen with underscores. 
For example,

command line option                 environment variable
------------------------------------------------------------------
--docker-sock-url                  	GOHORSE_DOCKER_SOCK_URL
--target.host.name               	GOHORSE_TARGET_HOST_NAME
	`,
	RunE: func(cmd *cobra.Command, args []string) error {

		filterBuilder := new(filterConfig.FilterBuilder).InitFromViper(viper.GetViper())
		filter := new(filters.FilterManager).InitFromFilterBuilder(filterBuilder)

		webBuilder := new(webConfig.WebBuilder).InitFromViper(viper.GetViper(), filter)
		server := new(web.Server).InitFromWebBuilder(webBuilder)

		filter.ListAPIs.Init()

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


