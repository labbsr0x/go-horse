package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/filters/list"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/handlers"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/model"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/sockClient"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var client = sockclient.Get("unix:///var/run/docker.sock")

func main() {
	app := iris.New()
	app.Use(recover.New())
	app.Post("/login", handlers.Login)
	app.Any("*", ProxyHandler)
	log.Warn().Msg("Inicializando o sandman-acl-proxy ... ")
	app.Run(iris.Addr(config.Port)) //:8080
}

func init() {
	zerolog.TimeFieldFormat = ""
}

// ProxyHandler lero-lero
func ProxyHandler(ctx iris.Context) {

	token, tokenlessURL, err := util.ExtractTokenFromURL(ctx)
	log.Info().Msg("TOKEN :> " + token)

	if err != nil {
		ctx.StatusCode(400)
		ctx.JSON(iris.Map{
			"message": err.Error(),
		})
		ctx.Next()
		return
	}
	fmt.Println("ctx.Request().URL.RawQuery : ", ctx.Request().URL.String())

	isAllowed := true //validatePolicy(ctx, tokenlessURL)

	if !isAllowed {
		ctx.StatusCode(403)
		ctx.JSON(iris.Map{
			"message": "Vivaldo disse : 'NÃO!'",
		})
		ctx.Next()
		return
	}

	if ctx.Values().Get("requestBody") == nil {
		requestBody, erro := ioutil.ReadAll(ctx.Request().Body)
		if erro != nil {
			fmt.Println("ERRO AO PARSEAR O BODY DO REQUEST PARA A REQUISICAO :: ", ctx.String())
		}
		ctx.Values().Set("requestBody", string(requestBody))
	}

	before(ctx)

	request, newRequestError := http.NewRequest(ctx.Request().Method, tokenlessURL, strings.NewReader(ctx.Values().GetString("requestBody")))

	for key, value := range ctx.Request().Header {
		request.Header[key] = value
	}

	if newRequestError != nil {
		fmt.Println("error new request ::>> ", newRequestError)
	}
	response, error := client.Do(request)

	if error != nil {
		fmt.Println("ERROR request docker sock :> ", error.Error())
		ctx.Next()
	}

	responseBody, erro := ioutil.ReadAll(response.Body)
	if erro != nil {
		ctx.WriteString("Erro parsear a resposta do token - " + erro.Error())
	}

	for key, value := range response.Header {
		ctx.Header(key, value[0])
	}

	ctx.Values().Set("responseBody", string(responseBody))

	after(ctx)

	ctx.ContentType("application/json")
	ctx.StatusCode(response.StatusCode)
	ctx.WriteString(ctx.Values().GetString("responseBody"))
}

func before(ctx iris.Context) {
	requestPath := ctx.Path()
	log.Info().Msg("BEFORE the mainHandler: " + requestPath)
	var requestBody []byte
	if ctx.Values().Get("requestBody") == nil {
		requestBody, erro := ioutil.ReadAll(ctx.Request().Body)
		if erro != nil {
			fmt.Println("ERRO AO PARSEAR O BODY DO REQUEST PARA A REQUISICAO :: ", ctx.String())
		}
		ctx.Values().Set("requestBody", string(requestBody))
	}

	for _, filter := range list.BeforeFilters() {
		if filter.MatchURL(ctx) {
			result := filter.Exec(ctx, string(requestBody))
			if result.Operation == model.Write {
				ctx.Values().Set("requestBody", result.Body)
			}
			if !result.Next {
				break
			}
		}
	}
}

func after(ctx iris.Context) {
	requestPath := ctx.Path()
	println("AFTER the mainHandler: " + requestPath)

	for _, filter := range list.AfterFilters() {
		if filter.MatchURL(ctx) {
			result := filter.Exec(ctx, ctx.Values().GetString("responseBody"))
			if result.Operation == model.Write {
				ctx.Values().Set("responseBody", result.Body)
			}
			if !result.Next {
				break
			}
		}
	}
}
