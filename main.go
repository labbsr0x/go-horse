package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/handlers"
	"github.com/kataras/iris"
)

var client = httpClient("unix:///var/run/docker.sock")

func main() {
	app := iris.Default()
	// app.Use(before)
	// app.Done(after)
	app.Post("/login", handlers.Login)
	app.Any("*", before, ProxyHandler, after)
	app.Run(iris.Addr(config.PORT))
}

func httpClient(u string) *http.Client {
	url, err := url.Parse(u)
	if err != nil {
		fmt.Println("failed parsing URL", u, " : ", err)
		return nil
	}
	transport := &http.Transport{}
	transport.DisableKeepAlives = true
	path := url.Path
	transport.Dial = func(proto, addr string) (net.Conn, error) {
		return net.Dial("unix", path)
	}
	url.Scheme = "http"
	url.Host = "unix-socket"
	url.Path = ""
	return &http.Client{Transport: transport}
}

// ProxyHandler lero-lero
func ProxyHandler(ctx iris.Context) {
	println("Inside mainHandler")
	// info := ctx.Values().GetString("info")
	request, newRequestError := http.NewRequest(ctx.Request().Method, "http://sandman"+ctx.Request().URL.String(), ctx.Request().Body)
	if newRequestError != nil {
		fmt.Println("erroe new request ::>> ", newRequestError)
	}
	response, error := client.Do(request)

	if error != nil {
		fmt.Println("ERROR request docker sock :> ", error)
		ctx.Next()
	}

	responseBody, erro := ioutil.ReadAll(response.Body)
	if erro != nil {
		ctx.WriteString("Erro parsear a resposta do token - " + erro.Error())
	}

	fmt.Println("response :: >> ", string(responseBody), error)
	ctx.ContentType("application/json")
	ctx.StatusCode(response.StatusCode)
	ctx.Write(responseBody)
	ctx.Next()
}

func before(ctx iris.Context) {
	shareInformation := "this is a sharable information between handlers"
	requestPath := ctx.Path()
	println("Before the mainHandler: " + requestPath)
	ctx.Values().Set("info", shareInformation)
	ctx.Next()
}

func after(ctx iris.Context) {
	println("After the mainHandler")
}
