package main

import (
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/server"
	"os"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/config"
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
	log.Info().Str("version", config.Version).Str("buildTime", config.BuildTime).Str("Commit", config.GitCommit).Msg("Version information")
	cmd.Execute()
}
