package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/handlers"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/matchers"
	"github.com/kataras/iris"
	"github.com/ory/ladon"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/filters"

	keto "github.com/ory/keto/sdk/go/keto/swagger"
)

var client = httpClient("unix:///var/run/docker.sock")

func main() {
	app := iris.Default()
	// app.Use(before)
	// app.Done(after)
	app.Post("/login", handlers.Login)
	app.Any("*", before, ProxyHandler, after)
	app.Run(iris.Addr(config.Port)) //:8080

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

func extractTokenFromURL(ctx iris.Context) (token, tokenlessURL string, err error) {
	url := ctx.Request().URL.String()
	fmt.Println("URL : ", url)
	paths := strings.Split(url, "/")
	if paths[1] != "token" || len(paths) < 4 {
		err = errors.New("URL inválida : verifique a variável de ambiente 'DOCKER_HOST' deve conter o host e o token no seguinte formato 'http://[host]/token/[token]'")
		return
	}
	token = paths[2]
	tokenlessURL = "http://sandman/" + strings.Join(paths[3:], "/")
	return
}

func validatePolicy(ctx iris.Context, tokenlessURL string) bool {
	// fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$ tokenlessURL ::::::>>>>>>> ", tokenlessURL)
	// ABILIO SAYS: extrair essa nojeira para um struct com paredes de chumbo para evitar vazamento e contaminação de todo o cluster com esse "shenanigan"
	dockerPath := strings.Split(strings.Join(strings.Split(tokenlessURL, "/")[4:], "/"), "?")[:1][0]
	if strings.Contains(tokenlessURL, "_ping") {
		return true
	}
	// fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$ DOCKERPATH ::::::>>>>>>> ", dockerPath)
	urlServer := "http://172.24.40.63:4466"
	// fmt.Println("Testing API...")
	client := keto.NewWardenApiWithBasePath(urlServer)
	result, _, err := client.IsSubjectAuthorized(keto.WardenSubjectAuthorizationRequest{
		Action:   strings.ToUpper(ctx.Request().Method),
		Resource: "srn:campus:docker:region1:sandman:dockerapi/" + dockerPath,
		Subject:  "docker",
		Context:  ladon.Context{},
	})
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	// fmt.Printf("Allowed: %t\n", result.Allowed)
	return result.Allowed
}

// ProxyHandler lero-lero
func ProxyHandler(ctx iris.Context) {

	// SAAPORRA

	// log.Printf("Creating JS interpreter")
	// js := otto.New()
	// js.Set("ctx", ctx)
	// js.Set("url", ctx.Request().URL.String())

	// js.Set("abort", func(call otto.FunctionCall) otto.Value { //
	// 	// fmt.Printf("Hello, %s.\n", call.Argument(0).String())
	// 	// return otto.Value{}
	// 	ctx.StatusCode(555)
	// 	ctx.WriteString("Erro : deu merda")
	// 	ctx.Next()
	// 	return otto.Value{}
	// })

	// value, err := js.Run(
	// 	`(
	// 	function(teste){abc = 2 + 2;
	// 	console.log("$$$$$$$$$$$$$$$$$$$ ", url)
	// 	console.log("++++++++++++++ The value of abc is " + abc); // 4
	// 	console.log("+_+_+_+ ", teste)
	// 	abort();
	// 	return false})(url)`)

	// if ret, err := value.ToBoolean(); err == nil && !ret {
	// 	return
	// }

	// SAPORRA

	// println("Inside mainHandler")
	// info := ctx.Values().GetString("info")
	token, tokenlessURL, err := extractTokenFromURL(ctx)
	fmt.Println("TOKEN :> ", token)
	// fmt.Println("URL SEM TOKEN :> ", tokenlessURL)
	// fmt.Println("ERR :> ", err)
	if err != nil {
		ctx.StatusCode(400)
		ctx.JSON(iris.Map{
			"message": err.Error(),
		})
		ctx.Next()
		return
	}
	fmt.Println("ctx.Request().URL.RawQuery : ", ctx.Request().URL.String())
	out, _ := httputil.DumpRequest(ctx.Request(), true)
	fmt.Println("$$$$$$$$$$$$$$$ >>>>>>>>>>>>>>>>>>> REQUEST : \n", string(out))

	isAllowed := true //validatePolicy(ctx, tokenlessURL)

	if !isAllowed {
		ctx.StatusCode(403)
		ctx.JSON(iris.Map{
			"message": "Vivaldo disse : 'NÃO!'",
		})
		ctx.Next()
		return
	}

	request, newRequestError := http.NewRequest(ctx.Request().Method, tokenlessURL, ctx.Request().Body)

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

	fmt.Println("$$$$$$$$$$$$$$$ >>>>>>>>>>>>>>>>>>> RESPONSE : \n", string(responseBody))

	// if strings.HasSuffix(ctx.Request().URL.String(), "/services/create") {
	// 	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	// 	var idResponse models.IDResponse
	// 	if err := ctx.ReadJSON(&idResponse); err != nil {
	// 		ctx.StatusCode(iris.StatusBadRequest)
	// 		ctx.WriteString(err.Error())
	// 		return
	// 	}
	// }

	// fmt.Println("response :: >> ", string(responseBody), error)

	jsonMatcher := matchers.JSONMatcher{Query: "Config.Labels", ExpectedValue: "2"}

	fmt.Println("################### jsonMatcher.Match()", jsonMatcher.Match(responseBody))

	for _, filter := range filters.FilterMoldels {
		fmt.Println(filter.Exec(ctx, string(responseBody)))
	}

	// ctx.Header("Api-Version", apiVersion)
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
