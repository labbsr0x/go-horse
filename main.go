package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/handlers"
	sockclient "gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/sockClient"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/util"
	"github.com/kataras/iris"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/filters"
)

var client *http.Client = sockclient.Get("unix:///var/run/docker.sock")

func main() {
	app := iris.Default()
	// app.Use(before)
	// app.Done(after)
	// teste := filters.JsFilterModel{Operation: filters.Read}
	// teste2 := filters.JsFilterModel{Operation: filters.Write}
	// fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ >>>>>>>>>>>>>>>>>>>>>>>> ", teste.Operation == filters.Read)
	// fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ >>>>>>>>>>>>>>>>>>>>>>>> ", teste2.Operation == filters.Read)
	// fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ >>>>>>>>>>>>>>>>>>>>>>>> ", teste2.Operation == filters.Write)
	app.Post("/login", handlers.Login)
	app.Any("*", before, ProxyHandler, after)
	app.Run(iris.Addr(config.Port)) //:8080
}

// ProxyHandler lero-lero
func ProxyHandler(ctx iris.Context) {

	// SAAPORRA QJ;QL;

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

	token, tokenlessURL, err := util.ExtractTokenFromURL(ctx)
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
	// out, _ := httputil.DumpRequest(ctx.Request(), true)
	// fmt.Println("$$$$$$$$$$$$$$$ >>>>>>>>>>>>>>>>>>> REQUEST : \n", string(out))

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

	// fmt.Println("$$$$$$$$$$$$$$$ >>>>>>>>>>>>>>>>>>> RESPONSE : \n", string(responseBody))

	ctx.Values().Set("responseBody", string(responseBody))

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

	// jsonMatcher := matchers.JSONMatcher{Query: "Config.Labels", ExpectedValue: "2"}

	// fmt.Println("################### jsonMatcher.Match()", jsonMatcher.Match(responseBody))

	// for _, filter := range filters.FilterMoldels {
	// 	result := filter.Exec(ctx, string(responseBody))
	// 	fmt.Printf("%+v\n", result)
	// 	fmt.Println("OPERATION :> ", result.Operation)
	// }

	// ctx.Header("Api-Version", apiVersion)
	ctx.ContentType("application/json")
	ctx.StatusCode(response.StatusCode)
	ctx.Write(responseBody)
	ctx.Next()
}

func before(ctx iris.Context) {
	requestPath := ctx.Path()
	println("BEFORE the mainHandler: " + requestPath)

	requestBody, erro := ioutil.ReadAll(ctx.Request().Body)
	if erro != nil {
		fmt.Println("ERRO AO PARSEAR O BODY DO REQUEST PARA A REQUISICAO :: ", ctx.String())
	}
	ctx.Values().Set("requestBody", string(requestBody))

	for _, filter := range filters.FiltersBefore {
		if filter.MatchURL(ctx) {
			result := filter.Exec(ctx, string(requestBody))
			fmt.Printf("%+v\n", result)
			fmt.Println("OPERATION :> ", result.Operation)
		}
	}

	ctx.Next()
}

func after(ctx iris.Context) {
	requestPath := ctx.Path()
	println("AFTER the mainHandler: " + requestPath)

	var responseBody string

	if strBody, ok := ctx.Values().Get("responseBody").(string); ok {
		responseBody = strBody
	} else {
		fmt.Println("ERRO AO PARSEAR O BODY DO REQUEST PARA A REQUISICAO :: ", strBody, ok)
	}

	for _, filter := range filters.FiltersAfter {
		if filter.MatchURL(ctx) {
			result := filter.Exec(ctx, responseBody)
			fmt.Printf("%+v\n", result)
			fmt.Println("OPERATION :> ", result.Operation)
		}
	}
}
