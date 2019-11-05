package list

import (
	"github.com/sirupsen/logrus"

	filter "github.com/labbsr0x/go-horse/filters/config-filter"
	"github.com/labbsr0x/go-horse/filters/filtergo"
	"github.com/labbsr0x/go-horse/filters/filterjs"

	"sort"
	"sync"
	"time"

	"github.com/labbsr0x/go-horse/plugins"

	"github.com/labbsr0x/go-horse/filters/model"
	"github.com/radovskyb/watcher"
)

// All requests and response filters
var all []model.Filter

// Request requests filters
var request []model.Filter

// Response response filters
var response []model.Filter

var updateLock = sync.WaitGroup{}
var isUpdating = false

type ListAPI interface {
	Load()
	createDirWatcher() *watcher.Watcher
	RequestFilters() []model.Filter
	ResponseFilters() []model.Filter
	Init()
	Reload()
}

type DefaultListAPI struct {
	*filter.FilterBuilder
}

func (dapi *DefaultListAPI) InitFromFilterBuilder(filterBuilder *filter.FilterBuilder) *DefaultListAPI {
	dapi.FilterBuilder = filterBuilder
	return dapi
}

// RequestFilters lero lero
func (dapi *DefaultListAPI) RequestFilters() []model.Filter {
	if isUpdating {
		updateLock.Wait()
	}
	return request
}

// ResponseFilters lero lero
func (dapi *DefaultListAPI) ResponseFilters() []model.Filter {
	if isUpdating {
		updateLock.Wait()
	}
	return response
}

func (dapi *DefaultListAPI) updateFilters() {
	updateLock.Add(1)
	isUpdating = true
	dapi.Load()
	updateLock.Done()
	isUpdating = false
}

func (dapi *DefaultListAPI) Init() {
	dapi.Reload()
}

// Reload Reload
func (dapi *DefaultListAPI) Reload() {
	dapi.createDirWatcher()
	dapi.Load()
}

// Load Load
func (dapi *DefaultListAPI) Load() {

	all = all[:0]
	request = request[:0]
	response = response[:0]

	jsFilters := filterjs.Load(dapi.FlagsFilter.JsFiltersPath)
	goFilters := plugins.Load(dapi.FlagsFilter.GoPluginsPath)

	for _, jsFilter := range jsFilters {
		filter := filterjs.NewFilterJS(jsFilter)
		all = append(all, filter)
		if filter.Config().Invoke == model.Request {
			request = append(request, filter)
		} else {
			response = append(response, filter)
		}
	}

	for _, goFilter := range goFilters {
		filter := filtergo.NewFilterGO(goFilter)
		all = append(all, filter)
		if filter.Config().Invoke == model.Request {
			request = append(request, filter)
		} else {
			response = append(response, filter)
		}
	}

	dapi.validateFilterOrder(request)
	dapi.validateFilterOrder(response)
	dapi.orderFilterModels(all, request, response)
}

func (dapi *DefaultListAPI) orderFilterModels(models ...[]model.Filter) {
	for _, filters := range models {
		sort.SliceStable(filters[:], func(i, j int) bool {
			return filters[i].Config().Order < filters[j].Config().Order
		})
	}
}

func (dapi *DefaultListAPI) validateFilterOrder(models []model.Filter) {
	last := -1
	for _, filter := range models {
		if filter.Config().Order == last {
			logrus.WithFields(logrus.Fields{
				"2 or more filters with the same order value": last,
			}).Fatalf("Error on filters definitions : property configuration mismatch ORDER")
			panic("Fix the filters configurations, including the plugins")
		}
		last = filter.Config().Order
	}
}

func (dapi *DefaultListAPI) createDirWatcher() *watcher.Watcher {

	var dirWatcher = watcher.New()

	go func() {
		for {
			select {
			case event := <-dirWatcher.Event:
				dapi.updateFilters()
				logrus.WithFields(logrus.Fields{
					"event": event,
				}).Warnf("Filters definition updated")
			case err := <-dirWatcher.Error:
				logrus.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Errorf("DirWatcher error")
			case <-dirWatcher.Closed:
				return
			}
		}
	}()

	if err := dirWatcher.AddRecursive(dapi.FlagsFilter.JsFiltersPath); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("DirWatcher error")
	}


	go func() {
		if err := dirWatcher.Start(time.Second); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Errorf("DirWatcher error")
		}
	}()



	return dirWatcher
}
