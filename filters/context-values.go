package filters

import (
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/kataras/iris"
	"github.com/robertkrimen/otto"
	"github.com/rs/zerolog/log"
)

func requestScopeGetToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	key, error := call.Argument(0).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing requestScopeGetToJSContext key field - js filter exec")
	}
	value := util.RequestScopeGet(ctx, key)
	result, error := otto.ToValue(value)
	if error != nil {
		log.Error().Err(error).Msg("Error parsing requestScopeGetToJSContext function return - js filter exec")
	}
	return result
}

func requestScopeSetToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	key, error := call.Argument(0).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing requestScopeSetToJSContext key field - js filter exec")
	}
	value, error := call.Argument(1).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing requestScopeSetToJSContext function exec - js filter exec")
	}
	util.RequestScopeSet(ctx, key, value)
	return otto.NullValue()
}

func requestScopeListToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value {
	mapa := util.RequestScopeList(ctx)
	result, error := call.Otto.ToValue(mapa)
	if error != nil {
		log.Error().Err(error).Msg("Error parsing requestScopeListToJSContext response map - js filter exec")
	}
	return result
}
