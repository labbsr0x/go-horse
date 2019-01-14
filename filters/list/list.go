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

// All lero lero
var all []model.Filter

// Request lero lero
var request []model.Filter

// Response lero lero
var response []model.Filter

var updateLock = sync.WaitGroup{}
var isUpdating = false

var dirWatcher *watcher.Watcher

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
	dirWatcher = createDirWatcher()
	Load()
}

// Load Load
func Load() {

	all = all[:0]
	request = request[:0]
	response = response[:0]

	jsFilters := filterjs.Load()
	goFilters := plugins.Load()
	for _, jsfilter := range jsFilters {
		filter := model.NewFilterJS(jsfilter)
		All = append(All, filter)
		if filter.Config().Invoke == model.Request {
			request = append(request, filter)
		} else {
			response = append(response, filter)
		}
	}
	for _, gofilter := range goFilters {
		filter := model.NewFilterGO(gofilter)
		All = append(All, filter)
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
	var watcher = watcher.New()

	go func() {
		for {
			select {
			case event := <-watcher.Event:
				updateFilters()
				log.Warn().Msg(fmt.Sprintf("Filters definition updated : %#v", event))
			case err := <-watcher.Error:
				log.Error().Err(err).Msg("DirWatcher error")
			case <-watcher.Closed:
				return
			}
		}
	}()

	if err := watcher.AddRecursive(config.JsFiltersPath); err != nil {
		log.Error().Err(err).Msg("DirWatcher error")
	}

	go func() {
		if err := watcher.Start(time.Second); err != nil {
			log.Error().Err(err).Msg("DirWatcher error")
		}
	}()

	return watcher
}
