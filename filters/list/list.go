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
var All []model.Filter

// Request lero lero
var Request []model.Filter

// Response lero lero
var Response []model.Filter

var updateLock = sync.WaitGroup{}
var isUpdating = false

// RequestFilters lero lero
func RequestFilters() []model.Filter {
	if isUpdating {
		updateLock.Wait()
	}
	return Request
}

// ResponseFilters lero lero
func ResponseFilters() []model.Filter {
	if isUpdating {
		updateLock.Wait()
	}
	return Response
}

func updateFilters() {
	updateLock.Add(1)
	isUpdating = true
	Load()
	updateLock.Done()
	isUpdating = false
}

func init() {
	dirWatcher()
	Load()
}

// Load lero-lero
func Load() {

	All = All[:0]
	Request = Request[:0]
	Response = Response[:0]

	jsFilters := filterjs.Load()
	goFilters := plugins.Load()
	for _, jsFilter := range jsFilters {
		filter := filterjs.NewFilterJS(jsFilter)
		All = append(All, filter)
		if filter.Config().Invoke == model.Request {
			Request = append(Request, filter)
		} else {
			Response = append(Response, filter)
		}
	}
	for _, goFilter := range goFilters {
		filter := filtergo.NewFilterGO(goFilter)
		All = append(All, filter)
		if filter.Config().Invoke == model.Request {
			Request = append(Request, filter)
		} else {
			Response = append(Response, filter)
		}
	}
	validateFilterOrder(Request)
	validateFilterOrder(Response)
	orderFilterModels(All, Request, Response)
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

func dirWatcher() {
	dirWatcher := watcher.New()

	go func() {
		for {
			select {
			case event := <-dirWatcher.Event:
				log.Warn().Msg(fmt.Sprintf("Filters definition updated : %#v", event))
				updateFilters()
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
}
