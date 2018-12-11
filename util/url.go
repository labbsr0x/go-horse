package util

import (
	"errors"
	"strings"

	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
)

// ExtractTokenFromURL lero-lero
func ExtractTokenFromURL(ctx iris.Context) (token, tokenlessURL string, err error) {
	url := ctx.Request().URL.String()
	log.Debug().Str("URL", url)
	paths := strings.Split(url, "/")
	if paths[1] != "token" || len(paths) < 4 {
		log.Warn().Msg("Error on url format from docker client : " + url)
		err = errors.New("URL inválida : verifique a variável de ambiente 'DOCKER_HOST' deve conter o host e o token no seguinte formato 'http://[host]/token/[token]'")
		return
	}
	token = paths[2]
	tokenlessURL = "http://sandman/" + strings.Join(paths[3:], "/")
	return
}
