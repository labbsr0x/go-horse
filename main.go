package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/filters/list"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/handlers"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/model"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/sockClient"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var client = sockclient.Get(config.DockerSockURL)

func main() {
	app := iris.New()
	app.Use(recover.New())
	app.Post("/login", handlers.Login)
	app.Any("*", ProxyHandler)
	app.Run(iris.Addr(config.Port))
}

func init() {
	// zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(config.LogLevel)
	if config.PrettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// panic (zerolog.PanicLevel, 5)
	// fatal (zerolog.FatalLevel, 4)
	// error (zerolog.ErrorLevel, 3)
	// warn (zerolog.WarnLevel, 2)
	// info (zerolog.InfoLevel, 1)
	// debug (zerolog.DebugLevel, 0)

}

// ProxyHandler lero-lero
func ProxyHandler(ctx iris.Context) {

	// token, tokenlessURL, err := util.ExtractTokenFromURL(ctx)
	log.Debug().Str("request", ctx.String())
	// log.Debug().Str("token", token)

	// if err != nil {
	// 	ctx.StatusCode(400)
	// 	ctx.JSON(iris.Map{"message": err.Error()})
	// 	ctx.Next()
	// 	return
	// }

	// isAllowed := security.VerifyPolicy(ctx.Request().Method, tokenlessURL)

	// if !isAllowed {
	// 	ctx.StatusCode(403)
	// 	ctx.JSON(iris.Map{
	// 		"message": "Vivaldo disse : 'NÃO!'",
	// 	})
	// 	ctx.Next()
	// 	return
	// }

	if ctx.Request().Body != nil {
		requestBody, erro := ioutil.ReadAll(ctx.Request().Body)
		if erro != nil {
			log.Error().Str("request", ctx.String()).Err(erro)
		}
		ctx.Values().Set("requestBody", string(requestBody))
	}

	filterReturn := before(ctx)

	if filterReturn.Err != nil {
		if filterReturn.Status == 0 {
			filterReturn.Status = http.StatusInternalServerError
		}
		ctx.StatusCode(filterReturn.Status)
		ctx.ContentType("application/json")
		ctx.WriteString(filterReturn.Err.Error())
		return
	}

	targetURL := ctx.Values().GetString("targetEndpoint")
	if targetURL == "" {
		targetURL = config.TargetHostname + ctx.Request().URL.RequestURI()
	}

	request, newRequestError := http.NewRequest(ctx.Request().Method, targetURL, strings.NewReader(ctx.Values().GetString("requestBody")))

	if newRequestError != nil {
		log.Error().Str("request", ctx.String()).Err(newRequestError).Msg("Error creating a new request in main handler")
	}

	for key, value := range ctx.Request().Header {
		request.Header[key] = value
	}

	log.Debug().Msg("Executing request for URL : " + targetURL + " ...")
	response, error := client.Do(request)

	if error != nil {
		log.Error().Str("request", ctx.String()).Err(error).Msg("Error executing the request in main handler")
		ctx.Next()
	}

	responseBody, erro := ioutil.ReadAll(response.Body)
	if erro != nil {
		ctx.WriteString("Erro parsear a resposta do token - " + erro.Error())
		log.Error().Str("request", ctx.String()).Err(erro).Msg("Error parsing response body in main handler")
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

func before(ctx iris.Context) model.FilterReturn {
	requestPath := ctx.Path()
	log.Debug().Msg("Before the mainHandler: " + requestPath)

	var result model.FilterReturn
	for _, filter := range list.BeforeFilters() {
		if filter.MatchURL(ctx) {
			log.Debug().Str("Filter matched : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", filter.Config()))
			result = filter.Exec(ctx, ctx.Values().GetString("requestBody"))
			log.Debug().Str("Filter output : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", result))
			if result.Operation == model.Write {
				log.Debug().Str("Body rewrite for filter - ", filter.Config().Name)
				ctx.Values().Set("requestBody", result.Body)
			}
			if !result.Next {
				log.Debug().Str("Filter chain canceled by filter - ", filter.Config().Name).Msg("lero-lero")
				break
			}
		}
	}
	return result
}

func after(ctx iris.Context) {
	requestPath := ctx.Path()
	log.Debug().Msg("After the mainHandler:" + requestPath)

	for _, filter := range list.AfterFilters() {
		if filter.MatchURL(ctx) {
			log.Debug().Str("Filter matched : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", filter.Config()))
			result := filter.Exec(ctx, ctx.Values().GetString("responseBody"))
			log.Debug().Str("Filter output : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", result))
			if result.Operation == model.Write {
				log.Debug().Str("Body rewrite for filter - ", filter.Config().Name)
				ctx.Values().Set("responseBody", result.Body)
			}
			if !result.Next {
				log.Debug().Str("Filter chain canceled by filter - ", filter.Config().Name)
				break
			}
		}
	}
}
