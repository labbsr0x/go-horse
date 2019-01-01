package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/config"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/list"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/model"
	sockclient "gitex.labbs.com.br/labbsr0x/proxy/go-horse/sockClient"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/docker/docker/client"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
)

var sockClient = sockclient.Get(config.DockerSockURL)
var dockerCli *client.Client

var waitChannel = make(chan int)

func init() {
	os.Setenv("DOCKER_API_VERSION", "1.39")
	os.Setenv("DOCKER_HOST", "unix:///var/run/docker.sock")
	var err error
	dockerCli, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
}

// ProxyHandler lero-lero
func ProxyHandler(ctx iris.Context) {

	log.Debug().Str("request", ctx.String()).Msg("Receiving request")

	util.SetEnvVars(ctx)

	if ctx.Request().Body != nil {
		requestBody, erro := ioutil.ReadAll(ctx.Request().Body)
		if erro != nil {
			log.Error().Str("request", ctx.String()).Err(erro)
		}
		ctx.Values().Set("requestBody", string(requestBody))
	}

	// mussum was here
	_, erris := runRequestFilters(ctx)

	if erris != nil {
		ctx.StopExecution()
		return
	}

	targetURL := ctx.Values().GetString("targetEndpoint")
	if targetURL == "" {
		targetURL = ctx.Request().URL.RequestURI()
	}

	request, newRequestError := http.NewRequest(ctx.Request().Method, config.TargetHostname+targetURL, strings.NewReader(ctx.Values().GetString("requestBody")))

	if newRequestError != nil {
		log.Error().Str("request", ctx.String()).Err(newRequestError).Msg("Error creating a new request in main handler")
	}

	for key, value := range ctx.Request().Header {
		request.Header[key] = value
	}

	log.Debug().Msg("Executing request for URL : " + targetURL + " ...")

	response, erre := sockClient.Do(request)

	if erre != nil {
		log.Error().Str("request", ctx.String()).Err(erre).Msg("Error executing the request in main handler")
		ctx.Next()
		return
	}

	defer response.Body.Close()

	if strings.Contains(targetURL, "build") {

		ctx.ContentType("application/json")
		ctx.Header("Transfer-Encoding", "chunked")

		ctx.StreamWriter(func(w io.Writer) bool {
			var buf = make([]byte, 1024)
			read, er := response.Body.Read(buf)
			if er != nil || er == io.EOF {
				return false
			}
			w.Write(buf[:read])
			return true
		})
		ctx.StopExecution()
		return

	}

	responseBody, erro := ioutil.ReadAll(response.Body)
	if erro != nil {
		ctx.WriteString("Error reading the response body - " + erro.Error())
		log.Error().Str("request", ctx.String()).Err(erro).Msg("Error parsing response body in main handler")
	}

	for key, value := range response.Header {
		ctx.Header(key, value[0])
	}

	ctx.Values().Set("responseBody", string(responseBody))

	errr := runResponseFilters(ctx)

	if errr != nil {
		log.Error().Err(errr).Msg("Error during the execution of response filters")
	}

	ctx.ContentType("application/json")
	ctx.StatusCode(response.StatusCode)
	ctx.WriteString(ctx.Values().GetString("responseBody"))

}

func runRequestFilters(ctx iris.Context) (result model.FilterReturn, err error) {
	requestPath := ctx.Path()
	log.Debug().Msg("Request the mainHandler: " + requestPath)

	for _, filter := range list.RequestFilters() {
		if filter.MatchURL(ctx) {
			log.Debug().Str("Filter matched : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", filter.Config()))
			result, err = filter.Exec(ctx, ctx.Values().GetString("requestBody"))
			log.Debug().Str("Filter output : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", result))
			if result.Operation == model.Write {
				log.Debug().Str("Body rewrite for filter - ", filter.Config().Name)
				ctx.Values().Set("requestBody", result.Body)
			}
			if !result.Next {
				log.Info().Str("Filter chain canceled by filter - ", filter.Config().Name).Msg("lero-lero")
				break
			}
		}
	}

	if err != nil {
		if result.Status == 0 {
			result.Status = http.StatusInternalServerError
		}
		ctx.StatusCode(result.Status)
		ctx.ContentType("application/json")
		ctx.WriteString(err.Error())
	}

	return
}

func runResponseFilters(ctx iris.Context) (err error) {
	requestPath := ctx.Path()
	log.Debug().Msg("Response the mainHandler:" + requestPath)

	for _, filter := range list.ResponseFilters() {
		if filter.MatchURL(ctx) {
			log.Debug().Str("Filter matched : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", filter.Config()))
			result, err := filter.Exec(ctx, ctx.Values().GetString("responseBody"))
			if err != nil {
				log.Error().Err(err).Msgf("Error applying filter : %s", filter.Config().Name)
			}
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

	return
}
