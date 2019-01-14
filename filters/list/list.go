package list

import (
	"fmt"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/filtergo"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/filterjs"

	"sort"
	"sync"
	"time"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/plugins"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/config"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/model"
	"github.com/radovskyb/watcher"
	"github.com/rs/zerolog/log"
)

// All requests and response filters
var all []model.Filter

// Request requests filters
var request []model.Filter

// Response response filters
var response []model.Filter

var updateLock = sync.WaitGroup{}
var isUpdating = false

// RequestFilters lero lero
func RequestFilters() []model.Filter {
	if isUpdating {
		updateLock.Wait()
	}
	return request
}

// ResponseFilters lero lero
func ResponseFilters() []model.Filter {
	if isUpdating {
		updateLock.Wait()
	}
	return response
}

func updateFilters() {
	updateLock.Add(1)
	isUpdating = true
	Load()
	updateLock.Done()
	isUpdating = false
}

func init() {
	Reload()
}

// Reload Reload
func Reload() {
	createDirWatcher()
	Load()
}

// Load Load
func Load() {

	all = all[:0]
	request = request[:0]
	response = response[:0]

	jsFilters := filterjs.Load()
	goFilters := plugins.Load()
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
	validateFilterOrder(request)
	validateFilterOrder(response)
	orderFilterModels(all, request, response)
}

func orderFilterModels(models ...[]model.Filter) {
	for _, filters := range models {
		sort.SliceStable(filters[:], func(i, j int) bool {
			return filters[i].Config().Order < filters[j].Config().Order
		})
	}
}

func validateFilterOrder(models []model.Filter) {
	last := -1
	for _, filter := range models {
		if filter.Config().Order == last {
			log.Fatal().Msg(fmt.Sprintf("Error on filters definitions : property configuration mismatch ORDER : theres 2 filters or more with the same order value -> %d", last))
			panic("Fix the filters configurations, including the plugins")
		}
		last = filter.Config().Order
	}
}

func createDirWatcher() *watcher.Watcher {
	var dirWatcher = watcher.New()

	go func() {
		for {
			select {
			case event := <-dirWatcher.Event:
				updateFilters()
				log.Warn().Msg(fmt.Sprintf("Filters definition updated : %#v", event))
			case err := <-dirWatcher.Error:
				log.Error().Err(err).Msg("DirWatcher error")
			case <-dirWatcher.Closed:
				return
			}
		}
	}()

	if err := dirWatcher.AddRecursive(config.JsFiltersPath); err != nil {
		log.Error().Err(err).Msg("DirWatcher error")
	}

	go func() {
		if err := dirWatcher.Start(time.Second); err != nil {
			log.Error().Err(err).Msg("DirWatcher error")
		}
	}()

	return dirWatcher
}
