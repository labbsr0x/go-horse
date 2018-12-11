package list

import (
	"fmt"

	"sort"
	"sync"
	"time"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/plugins"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/filters"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/model"
	"github.com/radovskyb/watcher"
	"github.com/rs/zerolog/log"
)

// All lero lero
var All []model.Filter

// Before lero lero
var Before []model.Filter

// After lero lero
var After []model.Filter

var updateLock = sync.WaitGroup{}
var isUpdating = false

// BeforeFilters lero lero
func BeforeFilters() []model.Filter {
	if isUpdating {
		updateLock.Wait()
	}
	return Before
}

// AfterFilters lero lero
func AfterFilters() []model.Filter {
	if isUpdating {
		updateLock.Wait()
	}
	return After
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
	Before = Before[:0]
	After = After[:0]

	jsFilters := filters.Load()
	goFilters := plugins.Load()
	for _, jsfilter := range jsFilters {
		filter := model.NewFilterJS(jsfilter)
		All = append(All, filter)
		if filter.Config().Invoke == model.Before {
			Before = append(Before, filter)
		} else {
			After = append(After, filter)
		}
	}
	for _, gofilter := range goFilters {
		filter := model.NewFilterGO(gofilter)
		All = append(All, filter)
		if filter.Config().Invoke == model.Before {
			Before = append(Before, filter)
		} else {
			After = append(After, filter)
		}
	}
	validateFilterOrder(All)
	orderFilterModels(All, Before, After)
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
			panic("Correct the filters configurations, including the plugins")
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
