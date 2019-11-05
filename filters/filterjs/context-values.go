package filterjs

import (
	"github.com/labbsr0x/go-horse/util"
	"github.com/kataras/iris"
	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
)

func requestScopeGetToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	key, err := call.Argument(0).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing requestScopeGetToJSContext key field - js filter exec")
	}
	value := util.RequestScopeGet(ctx, key)
	result, err := otto.ToValue(value)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing requestScopeGetToJSContext function return - js filter exec")
	}
	return result
}

func requestScopeSetToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	key, err := call.Argument(0).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing requestScopeSetToJSContext key field - js filter exec")
	}
	value, err := call.Argument(1).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing requestScopeSetToJSContext function exec - js filter exec")
	}
	util.RequestScopeSet(ctx, key, value)
	return otto.NullValue()
}

func requestScopeListToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	mapa := util.RequestScopeList(ctx)
	result, err := call.Otto.ToValue(mapa)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing requestScopeListToJSContext response map - js filter exec")
	}
	return result
}
