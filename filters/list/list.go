package list

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/plugins"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/filters"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/model"
	"github.com/radovskyb/watcher"
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
		fmt.Println("================================================>>>>>>>>>>>>>>>>>>>>>>>>>>> ", fmt.Sprintf("%#v", filter))
		All = append(All, filter)
		fmt.Println("================================================>>>>>>>>>>>>>>>>>>>>>>>>>>> ", fmt.Sprintf("%#v", filter.Config()))
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
		fmt.Println("order : ", filter.Config().Order)
		fmt.Println("last : ", last)
		if filter.Config().Order == last {
			panic(fmt.Sprintf("Erro na definição dos filtros : colisão da propriedade ordem : existem 2 filtros com a ordem nro -> %d", last))
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
				fmt.Println(event) // Print the event's info.
				updateFilters()
			case err := <-dirWatcher.Error:
				log.Fatalln("\n\n########### " + err.Error() + " ###########\n\n")
			case <-dirWatcher.Closed:
				return
			}
		}
	}()

	if err := dirWatcher.AddRecursive(config.JsFiltersPath); err != nil {
		log.Fatalln(err)
	}

	// if err := dirWatcher.AddRecursive(config.GoPluginsPath); err != nil {
	// 	log.Fatalln(err)
	// }

	go func() {
		if err := dirWatcher.Start(time.Second); err != nil {
			log.Fatalln(err)
		}
	}()
}
