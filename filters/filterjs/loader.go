package filterjs

import (
	"io/ioutil"
	"regexp"
	"strconv"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/model"
	"github.com/robertkrimen/otto"
	"github.com/rs/zerolog/log"
)

// Load load the filter from files
func Load(jsFiltersPath string) []model.FilterConfig {
	return parseFilterObject(readFromFile(jsFiltersPath))
}

func readFromFile(jsFiltersPath string) map[string]string {
	var jsFilterFunctions = make(map[string]string)

	files, err := ioutil.ReadDir(jsFiltersPath)
	if err != nil {
		log.Error().Err(err).Msg("Error reading filters dir - readFromFile")
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(jsFiltersPath + "/" + file.Name())
		if err != nil {
			log.Error().Err(err).Str("file", file.Name()).Msg("Error reading filter filter - readFromFile")
			continue
		}
		jsFilterFunctions[file.Name()] = string(content)
		log.Debug().Str("file", file.Name()).Str("filter_content", string(content)).Msg("js filter - readFromFile")
	}

	return jsFilterFunctions
}

func parseFilterObject(jsFilterFunctions map[string]string) []model.FilterConfig {
	var filterModels []model.FilterConfig

	fileNamePattern := regexp.MustCompile("^([0-9]{1,3})\\.(request|response)\\.(.*?)\\.js$")

	for fileName, jsFunc := range jsFilterFunctions {
		nameProperties := fileNamePattern.FindStringSubmatch(fileName)
		if nameProperties == nil || len(nameProperties) < 4 {
			log.Error().Str("file", fileName).Msg("Error file name")
			continue
		}

		order := nameProperties[1]
		invokeTime := nameProperties[2]
		name := nameProperties[3]

		js := otto.New()

		funcFilterDefinition, err := js.Call("(function(){return"+jsFunc+"})", nil, nil)
		if err != nil {
			log.Error().Err(err).Str("file", fileName).Msg("Error on JS object definition - parseFilterObject")
			continue
		}

		filter := funcFilterDefinition.Object()

		filterDefinition := model.FilterConfig{}

		if invokeTime == "request" {
			filterDefinition.Invoke = model.Request
		} else {
			filterDefinition.Invoke = model.Response
		}

		oderInt, orderParserError := strconv.Atoi(order)
		if orderParserError != nil {
			log.Error().Err(err).Str("file", fileName).Msg("Error on order int conversion - parseFilterObject")
			continue
		}
		filterDefinition.Order = oderInt
		filterDefinition.Name = name

		if value, err := filter.Get("pathPattern"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.PathPattern = value
				filterDefinition.Regex, err = regexp.Compile(value)
				if err != nil {
					log.Error().Str("plugin_name", filterDefinition.Name).Err(err).Msg("Error compiling the filter url matcher regex")
				}
			} else {
				log.Error().Err(err).Str("file", fileName).Str("field", "pathPattern").Msg("Error on JS filter definition - parseFilterObject")
			}
		} else {
			log.Error().Err(err).Str("file", fileName).Str("field", "pathPattern").Msg("Error on JS filter definition - parseFilterObject")
		}

		if value, err := filter.Get("function"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.Function = value
			} else {
				log.Error().Err(err).Str("file", fileName).Str("field", "function").Msg("Error on JS filter definition - parseFilterObject")
			}
		} else {
			log.Error().Err(err).Str("file", fileName).Str("field", "function").Msg("Error on JS filter definition - parseFilterObject")
		}

		filterModels = append(filterModels, filterDefinition)
	}
	return filterModels
}
