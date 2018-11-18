package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

// Login lero-lero BODY : {"username":"wsilva", "password":"12345678test"}
func Login(ctx iris.Context) {
	response, err := http.Post("http://172.24.40.16:3000/token/", context.ContentJSONHeaderValue, ctx.Request().Body)
	if err != nil {
		ctx.WriteString("Erro ao gerar o token - " + err.Error())
		fmt.Println(err.Error())
		return
	}
	responseBody, erro := ioutil.ReadAll(response.Body)
	if erro != nil {
		ctx.WriteString("Erro parsear a resposta do token - " + err.Error())
	}
	ctx.JSON(json.RawMessage(responseBody))
}
