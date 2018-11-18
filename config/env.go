package config

import (
	"os"
	"strings"
)

// PORT lero lero
var PORT = fixPortValue(os.Getenv("PORT"))

func fixPortValue(port string) string {
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}
