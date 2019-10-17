package util

import "github.com/rs/zerolog"

func ParseLogLevel(logLevel string) zerolog.Level {
	level, error := zerolog.ParseLevel(logLevel)
	if error != nil {
		return zerolog.WarnLevel
	}
	return level
}

