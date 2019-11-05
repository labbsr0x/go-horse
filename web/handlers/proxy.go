package handlers

import (
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/model"
	web "gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config-web"

	"github.com/kataras/iris"
)

const (
	RequestBodyKey  = "requestBody"
	ResponseBodyKey = "responseBody"
)

type ProxyAPI interface {
	ProxyHandler(ctx iris.Context)
}

type DefaultProxyAPI struct {
	*web.WebBuilder
}

func (dapi *DefaultProxyAPI) InitFromWebBuilder(webBuilder *web.WebBuilder) *DefaultProxyAPI {
	dapi.WebBuilder = webBuilder
	return dapi
}

func (dapi *DefaultProxyAPI) ProxyHandler(ctx iris.Context) {
	logrus.WithFields(logrus.Fields{
		"request": ctx.String(),
	}).Debugf("Receiving")

	u := ctx.Request().URL.ResolveReference(&url.URL{Path: ctx.Values().GetString("path"), RawQuery: ctx.Request().URL.RawQuery})
	path := u.String()

	request, newRequestError := http.NewRequest(ctx.Request().Method, dapi.Flags.TargetHostName+path, strings.NewReader(ctx.Values().GetString(RequestBodyKey)))

	if newRequestError != nil {
		logrus.WithFields(logrus.Fields{
			"request": ctx.String(),
			"error": newRequestError.Error(),
		}).Errorf("Error creating a new request in main handler")
	}

	for key, value := range ctx.Request().Header {
		request.Header[key] = value
	}

	logrus.WithFields(logrus.Fields{
		"URL": path,
	}).Debugf("Executing request for URL")

	response, err := dapi.SockClient.Do(request)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": ctx.String(),
			"error": err.Error(),
		}).Errorf("Error executing the request in main handle")
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

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.WriteString("Error reading the response body - " + err.Error())
		logrus.WithFields(logrus.Fields{
			"request": ctx.String(),
			"error": err.Error(),
		}).Errorf("Error parsing response body in main handler")
	}

	for key, value := range response.Header {
		if key != "Content-Length" {
			ctx.Header(key, value[0])
		}
	}

	ctx.Values().Set(ResponseBodyKey, string(responseBody))
	ctx.Values().Set("responseStatusCode", response.StatusCode)

	result, errr := dapi.Filter.RunResponseFilters(ctx, ResponseBodyKey)

	if errr != nil {
		logrus.WithFields(logrus.Fields{
			"error": errr.Error(),
		}).Errorf("Error during the execution of RESPONSE filters")
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
