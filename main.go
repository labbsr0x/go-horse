package main

import (
	"os"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/cmd"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/version"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(config.LogLevel)
	if config.PrettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func main() {
	log.Info().Str("version", version.Version).Str("buildTime", version.BuildTime).Str("Commit", version.GitCommit).Msg("Version information")
	cmd.Execute()
}
