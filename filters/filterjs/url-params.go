package filterjs

import (
	"github.com/kataras/iris"
	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
	"strings"
)


func listURLParamsFromJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	values := make(map[string]string)
	for key, value := range ctx.Request().URL.Query() {
		values[key] = strings.Join(value, ",")
	}

	result, err := call.Otto.ToValue(values)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing url-params map - js filter exec listToJSContext")
	}
	return result
}

func getURLParamsFromJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	key, err := call.Argument(0).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing addURLParamsFromJSContext key field - js filter exec")
	}

	for k, v := range ctx.Request().URL.Query() {
		if k == key {
			result, err := otto.ToValue(strings.Join(v, ","))
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Errorf("Error parsing requestScopeGetToJSContext function return - js filter exec")
			}
			return result
		}
	}
	return otto.NullValue()
}

func delURLParamsFromJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	key, err := call.Argument(0).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing addURLParamsFromJSContext key field - js filter exec")
	}
	ctx.Request().URL.Query().Del(key)
	return otto.NullValue()
}

func addURLParamsFromJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	key, err := call.Argument(0).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing addURLParamsFromJSContext key field - js filter exec")
	}
	value, err := call.Argument(1).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing addURLParamsFromJSContext value field - js filter exec")
	}

	q := ctx.Request().URL.Query()
	q.Add(key, value)

	ctx.Request().URL.RawQuery = q.Encode()
	return otto.NullValue()
}

func setURLParamsFromJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	key, err := call.Argument(0).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing addURLParamsFromJSContext key field - js filter exec")
	}
	value, err := call.Argument(1).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing addURLParamsFromJSContext value field - js filter exec")
	}

	q := ctx.Request().URL.Query()
	q.Set(key, value)

	ctx.Request().URL.RawQuery = q.Encode()
	return otto.NullValue()
}
