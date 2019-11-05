package filterjs

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/labbsr0x/go-horse/filters/model"
	"github.com/robertkrimen/otto"
)

// Load load the filter from files
func Load(jsFiltersPath string) []model.FilterConfig {
	return parseFilterObject(readFromFile(jsFiltersPath))
}

func readFromFile(jsFiltersPath string) map[string]string {
	var jsFilterFunctions = make(map[string]string)

	files, err := ioutil.ReadDir(jsFiltersPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error reading filters dir - readFromFile")
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(jsFiltersPath + "/" + file.Name())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"file name" : file.Name(),
				"error": err.Error(),
			}).Errorf("Error reading filter filter - readFromFile")
			continue
		}
		jsFilterFunctions[file.Name()] = string(content)


		logrus.WithFields(logrus.Fields{
			"file": file.Name(),
			"filter_content": string(content),
		}).Debugf("js filter - readFromFile")
	}

	return jsFilterFunctions
}

func parseFilterObject(jsFilterFunctions map[string]string) []model.FilterConfig {
	var filterModels []model.FilterConfig

	fileNamePattern := regexp.MustCompile("^([0-9]{1,3})\\.(request|response)\\.(.*?)\\.js$")

	for fileName, jsFunc := range jsFilterFunctions {

		logrusFileField := logrus.Fields{"file": fileName}

		nameProperties := fileNamePattern.FindStringSubmatch(fileName)
		if nameProperties == nil || len(nameProperties) < 4 {
			logrus.WithFields(logrusFileField).Errorf("Error file name")
			continue
		}

		order := nameProperties[1]
		invokeTime := nameProperties[2]
		name := nameProperties[3]

		js := otto.New()

		funcFilterDefinition, err := js.Call("(function(){return"+jsFunc+"})", nil, nil)
		if err != nil {
			logrus.WithFields(logrusFileField).Errorf("Error on JS object definition - parseFilterObject")
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
			logrus.WithFields(logrusFileField).Errorf("Error on order int conversion - parseFilterObject")
			continue
		}
		filterDefinition.Order = oderInt
		filterDefinition.Name = name

		if value, err := filter.Get("pathPattern"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.PathPattern = value
				filterDefinition.Regex, err = regexp.Compile(value)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"plugin_name": filterDefinition.Name,
						"error": err.Error(),
					}).Errorf("Error compiling the filter url matcher regex")
				}
			} else {
				logrus.WithFields(logrus.Fields{
					"file": fileName,
					"field": "pathPattern",
					"error": err.Error(),
				}).Errorf("Error on JS filter definition - parseFilterObject")
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"file": fileName,
				"field": "pathPattern",
				"error": err.Error(),
			}).Errorf("Error on JS filter definition - parseFilterObject")
		}

		if value, err := filter.Get("function"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.Function = value
			} else {
				logrus.WithFields(logrus.Fields{
					"file": fileName,
					"field": "function",
					"error": err.Error(),
				}).Errorf("Error on JS filter definition - parseFilterObject")
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"file": fileName,
				"field": "function",
				"error": err.Error(),
			}).Errorf("Error on JS filter definition - parseFilterObject")
		}

		filterModels = append(filterModels, filterDefinition)
	}
	return filterModels
}
