package data

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/alexflint/go-restructure"
	"github.com/wk8/go-ordered-map/v2"
)

type SortDir bool
const SortAsc SortDir = true
const SortDesc SortDir = false
func (dir SortDir) String() string {
	if dir == SortAsc {
		return "asc"
	}
	return "desc"
}

type SearchMap map[string]string
type InMap map[string]string
type SortMap orderedmap.OrderedMap[string, SortDir]

func NewSortMap() *SortMap {
	omap := orderedmap.New[string, SortDir]()
	return (*SortMap)(omap)
}

func (sm *SortMap) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	for k, v := range sm.FromOldest() {
		builder.WriteString(fmt.Sprintf("\"%s\":%v,", k, v))
	}
	builder.WriteString("}")
	return builder.String()
}

func (sm *SortMap) Get(key string) (SortDir, bool) {
	om := (*orderedmap.OrderedMap[string, SortDir])(sm)
	return om.Get(key)
}

func (sm *SortMap) Set(key string, dir SortDir) {
	om := (*orderedmap.OrderedMap[string, SortDir])(sm)
	om.Set(key, dir)
}

func (sm *SortMap) FromOldest() iter.Seq2[string, SortDir] {
	return func(yield func(string, SortDir) bool) {
		om := (*orderedmap.OrderedMap[string, SortDir])(sm)
		for pair := om.Oldest(); pair != nil; pair = pair.Next() {
			if !yield(pair.Key, pair.Value) {
				return
			}
		}
	}
}

type Filters struct {
	Search		*SearchMap
	Sort		*SortMap
	In			*InMap
	Page		*int
	PageSize	*int
}

type FilterConstraints struct {
	Search		func(SearchMap) bool
	Sort		func(SortMap) bool
	In			func(InMap) bool
	Page		func(int) bool
	PageSize	func(int) bool
}

type SearchFields []string
type SortFields []string
type InFields []string

type SearchMatch struct {
	_	struct{}	`^search_`
	Key	string		`\w+`
	_	struct{}	`$`
}
var SearchRex = restructure.MustCompile(
	SearchMatch{},
	restructure.Options{},
)

type InMatch struct {
	_	struct{}	`^in_`
	Key	string		`\w+`
	_	struct{}	`$`
}
var InRex = restructure.MustCompile(
	InMatch{},
	restructure.Options{},
)

type SortMatch struct {
	_	struct{}	`^`
	Dir	string		`[+-]?`
	Key	string		`\w+`
	_	struct{}	`$`
}
var SortRex = restructure.MustCompile(
	SortMatch{},
	restructure.Options{},
)


func (searchFields SearchFields) Check(searchMap SearchMap) bool {
	for key := range searchMap {
		if !slices.Contains(searchFields, key) {
			return false
		}
	}
	return true
}

func (nf InFields) Check(nm InMap) bool {
	for k := range nm {
		if !slices.Contains(nf, k) {
			return false
		}
	}
	return true
}

func (sortFields SortFields) Check(sortMap SortMap) bool {
	for k, _ := range sortMap.FromOldest() {
		if !slices.Contains(sortFields, k) {
			return false
		}
	}
	return true
}

func NewSearchFields(s ...string) SearchFields {
	return SearchFields(s)
}

func NewSortFields(s ...string) SortFields {
	return SortFields(s)
}

func NewInFields(s ...string) InFields {
	return InFields(s)
}
