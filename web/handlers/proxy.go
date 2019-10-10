package handlers

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/model"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config"

	sockclient "gitex.labbs.com.br/labbsr0x/proxy/go-horse/sockClient"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/docker/docker/client"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
)

const (
	RequestBodyKey  = "requestBody"
	ResponseBodyKey = "responseBody"
)

var sockClient = sockclient.Get(config.DockerSockURL)
var dockerCli *client.Client

type ProxyAPI interface {
	ProxyHandler(ctx iris.Context)
}

type DefaultProxyAPI struct {
	*config.WebBuilder
}

// InitFromWebBuilder initializes a default consent api instance from a web builder instance
func (dapi *DefaultProxyAPI) InitFromWebBuilder(webBuilder *config.WebBuilder) *DefaultProxyAPI {
	dapi.WebBuilder = webBuilder
	return dapi
}

func init() {
	var err error
	dockerCli, err = client.NewClientWithOpts(client.WithVersion(config.DockerAPIVersion), client.WithHost(config.DockerSockURL))
	if err != nil {
		panic(err)
	}
}

// ProxyHandler lero-lero
func (dapi *DefaultProxyAPI) ProxyHandler(ctx iris.Context) {

	log.Info().Str("request", ctx.String()).Msg("Receiving")

	util.SetFilterContextValues(ctx)

	if ctx.Request().Body != nil {
		requestBody, erro := ioutil.ReadAll(ctx.Request().Body)
		if erro != nil {
			log.Error().Str("request", ctx.String()).Err(erro)
		}
		ctx.Values().Set(RequestBodyKey, string(requestBody))
	}

	ctx.Values().Set("path", ctx.Request().URL.Path)

	// mussum was here
	_, erris := filters.RunRequestFilters(ctx, RequestBodyKey)

	if erris != nil {
		log.Error().Err(erris).Msg("Error during the execution of REQUEST filters")
		ctx.StopExecution()
		return
	}

	u := ctx.Request().URL.ResolveReference(&url.URL{Path: ctx.Values().GetString("path"), RawQuery: ctx.Request().URL.RawQuery})
	path := u.String()

	request, newRequestError := http.NewRequest(ctx.Request().Method, dapi.Flags.TargetHostName+path, strings.NewReader(ctx.Values().GetString(RequestBodyKey)))

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

	ctx.Values().Set(ResponseBodyKey, string(responseBody))
	ctx.Values().Set("responseStatusCode", response.StatusCode)

	result, errr := filters.RunResponseFilters(ctx, ResponseBodyKey)

	if errr != nil {
		log.Error().Err(errr).Msg("Error during the execution of RESPONSE filters")
		ctx.StopExecution()
		return
	}

	ctx.StatusCode(fixZeroStatus(result, response))
	ctx.ContentType("application/json")
	ctx.WriteString(ctx.Values().GetString(ResponseBodyKey))
}

func fixZeroStatus(result model.FilterReturn, response *http.Response) int {
	if result.Status == 0 {
		return response.StatusCode
	}
	return result.Status
}
