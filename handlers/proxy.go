package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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

func init() {
	var err error
	dockerCli, err = client.NewClientWithOpts(client.WithVersion(config.DockerAPIVersion), client.WithHost(config.DockerSockURL))
	if err != nil {
		panic(err)
	}
}

// ProxyHandler lero-lero
func ProxyHandler(ctx iris.Context) {

	log.Info().Str("request", ctx.String()).Msg("Receiving request")

	util.SetEnvVars(ctx)

	if ctx.Request().Body != nil {
		requestBody, erro := ioutil.ReadAll(ctx.Request().Body)
		if erro != nil {
			log.Error().Str("request", ctx.String()).Err(erro)
		}
		ctx.Values().Set("requestBody", string(requestBody))
	}

	ctx.Values().Set("path", ctx.Request().URL.Path)

	// mussum was here
	_, erris := runRequestFilters(ctx)

	if erris != nil {
		log.Error().Err(erris).Msg("Error during the execution of REQUEST filters")
		ctx.StopExecution()
		return
	}

	u := ctx.Request().URL.ResolveReference(&url.URL{Path: ctx.Values().GetString("path"), RawQuery: ctx.Request().URL.RawQuery})
	path := u.String()

	request, newRequestError := http.NewRequest(ctx.Request().Method, config.TargetHostname+path, strings.NewReader(ctx.Values().GetString("requestBody")))

	if newRequestError != nil {
		log.Error().Str("request", ctx.String()).Err(newRequestError).Msg("Error creating a new request in main handler")
	}

	for key, value := range ctx.Request().Header {
		request.Header[key] = value
	}

	log.Debug().Msg("Executing request for URL : " + path + " ...")

	response, erre := sockClient.Do(request)

	if erre != nil {
		log.Error().Str("request", ctx.String()).Err(erre).Msg("Error executing the request in main handler")
		ctx.Next()
		return
	}

	defer response.Body.Close()

	if strings.Contains(path, "build") {

		ctx.ResetResponseWriter(ctx.ResponseWriter())
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
		if key != "Content-Length" {
			ctx.Header(key, value[0])
		}
	}

	ctx.Values().Set("responseBody", string(responseBody))
	ctx.Values().Set("responseStatusCode", response.StatusCode)

	result, errr := runResponseFilters(ctx)

	if errr != nil {
		log.Error().Err(errr).Msg("Error during the execution of RESPONSE filters")
		ctx.StopExecution()
		return
	}

	ctx.StatusCode(fixZeroStatus(result, response))
	ctx.ContentType("application/json")
	ctx.WriteString(ctx.Values().GetString("responseBody"))
	fmt.Println(ctx.Values().GetString("responseBody"))
}

func fixZeroStatus(result model.FilterReturn, response *http.Response) int {
	if result.Status == 0 {
		return response.StatusCode
	}
	return result.Status
}

func runRequestFilters(ctx iris.Context) (result model.FilterReturn, err error) {
	requestPath := ctx.Path()
	log.Debug().Msgf("Running REQUEST filters for url : %s", requestPath)

	for _, filter := range list.RequestFilters() {
		if filter.MatchURL(ctx) {
			log.Debug().Str("Filter matched", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", filter.Config())).Msg("executing filter ...")
			result, err = filter.Exec(ctx, ctx.Values().GetString("requestBody"))
			if err != nil {
				log.Error().Err(err).Msgf("Error applying filter : %s", filter.Config().Name)
			}
			log.Debug().Str("Filter output", fmt.Sprintf("%#v", result)).Str("filter_config", fmt.Sprintf("%#v", result)).Msg("filter execution end")
			if result.Operation == model.Write {
				log.Debug().Msgf("Body rewrite for filter : %s", filter.Config().Name)
				ctx.Values().Set("requestBody", result.Body)
			}
			if !result.Next {
				log.Info().Msgf("Filter chain canceled by filter - %s", filter.Config().Name)
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

func runResponseFilters(ctx iris.Context) (result model.FilterReturn, err error) {

	requestPath := ctx.Path()
	log.Debug().Msgf("Running RESPONSE filters for url : %s", requestPath)

	for _, filter := range list.ResponseFilters() {
		if filter.MatchURL(ctx) {
			log.Debug().Str("Filter matched", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", filter.Config())).Msg("executing filter ...")
			result, err = filter.Exec(ctx, ctx.Values().GetString("responseBody"))
			if err != nil {
				log.Error().Err(err).Msgf("Error applying filter : %s", filter.Config().Name)
			}
			log.Debug().Str("Filter output", fmt.Sprintf("%#v", result)).Str("filter_config", fmt.Sprintf("%#v", result)).Msg("filter execution end")
			if result.Operation == model.Write {
				log.Debug().Msgf("Body rewrite for filter : %s", filter.Config().Name)
				ctx.Values().Set("responseBody", result.Body)
				ctx.StatusCode(result.Status)
			}
			if !result.Next {
				log.Info().Msgf("Filter chain canceled by filter - %s", filter.Config().Name)
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
