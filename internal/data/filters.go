package data

import (
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/alexflint/go-restructure"
	"github.com/dusktreader/the-hunt/internal/validator"
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

// Thanks to: https://betterstack.com/community/guides/logging/logging-in-go/
func (sm *SortMap) LogValue() slog.Value {
	args := make([]slog.Attr, 0, 5)
	for k, v := range sm.FromOldest() {
		args = append(args, slog.String(k, v.String()))
	}
	return slog.GroupValue(args...)

}

func readString(
	qs url.Values,
	key string,
	_ *validator.Validator,
) *string {
	var p *string
	s := qs.Get(key)
	if s != "" {
		slog.Debug("Read string", "key", key, "value", s)
		p = &s
	}
	return p
}

func readCSV(
	qs url.Values,
	key string,
	_ *validator.Validator,
) []string {
	s := make([]string, 0)
	csv := qs.Get(key)
	if csv != "" {
		s = strings.Split(csv, ",")
	}
	return s
}

func readInt(
	qs url.Values,
	key string,
	v *validator.Validator,
) *int {
	var p *int
	s := qs.Get(key)
	if s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			v.AddError("query", "must be an integer")
		} else {
			p = &i
		}
	}
	return p
}

/* Keep this around if we need it, but not used ATM
func readBool(
	qs url.Values,
	key string,
	v *validator.Validator,
) *bool {
	var p *bool
	s := qs.Get(key)
	if s != "" {
		sl := strings.ToLower(s)
		if slices.Contains([]string{"t", "true", "y", "yes", "1"}, sl) {
			b := true
			p = &b
		} else if slices.Contains([]string{"f", "false", "n", "no", "0"}, sl) {
			b := false
			p = &b
		} else {
			v.AddError("query", fmt.Sprintf("could not map %q to a boolean value", s))
		}
	}
	return p
}
*/

const DefaultPage = 1
const DefaultMaxPage = 1
const DefaultPageSize = 10
const DefaultMaxPageSize = 10

type Filters struct {
	Search   *SearchMap
	Sort     *SortMap
	In       *InMap
	Page     *int
	PageSize *int
}

func (f Filters) LogValue() slog.Value {
	args := make([]slog.Attr, 0, 5)
	if f.Search != nil {
		args = append(args, slog.Any("search", f.Search))
	}
	if f.Sort != nil {
		args = append(args, slog.Any("sort", f.Sort))
	}
	if f.In != nil {
		args = append(args, slog.Any("in", f.In))
	}
	if f.Page != nil {
		args = append(args, slog.Int("page", *f.Page))
	}
	if f.PageSize != nil {
		args = append(args, slog.Int("page_size", *f.PageSize))
	}

	return slog.GroupValue(args...)
}

type SearchCheck func(SearchMap) bool
type SortCheck func(SortMap) bool
type InCheck func(InMap) bool
type PageCheck func(int) bool
type PageSizeCheck func(int) bool

func DefaultPageCheck(i int) bool {
	return i >= 1 || i <= DefaultMaxPage
}

func DefaultPageSizeCheck(i int) bool {
	return i >= 1 || i <= DefaultMaxPageSize
}

type FilterConstraints struct {
	Search   SearchCheck
	Sort     SortCheck
	In       InCheck
	Page     PageCheck
	PageSize PageSizeCheck
}

type SearchFields []string
type SortFields []string
type InFields []string

type SearchMatch struct {
	_   struct{} `regexp:"^search_"`
	Key string   `regexp:"\\w+"`
	_   struct{} `regexp:"$"`
}

var SearchRex = restructure.MustCompile(
	SearchMatch{},
	restructure.Options{},
)

type InMatch struct {
	_   struct{} `regexp:"^in_"`
	Key string   `regexp:"\\w+"`
	_   struct{} `regexp:"$"`
}

var InRex = restructure.MustCompile(
	InMatch{},
	restructure.Options{},
)

type SortMatch struct {
	_   struct{} `regexp:"^"`
	Dir string   `regexp:"-?"`
	Key string   `regexp:"\\w+"`
	_   struct{} `regexp:"$"`
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

func ParseFilters(
	qs url.Values,
	v *validator.Validator,
	c FilterConstraints,
) Filters {
	f := Filters{
		Search:   &SearchMap{},
		In:       &InMap{},
		Page:     readInt(qs, "page", v),
		PageSize: readInt(qs, "page_size", v),
		Sort:     NewSortMap(),
	}

	var search SearchMatch
	var in InMatch
	for key := range qs {
		if SearchRex.Find(&search, key) {
			partKey := search.Key
			partVal := *readString(qs, key, v)
			if len(partVal) < 3 {
				v.AddError(key, "Search parameters must be at least 3 characters long")
			} else {
				(*f.Search)[partKey] = partVal
			}
		}

		if InRex.Find(&in, key) {
			partKey := in.Key
			partVal := *readString(qs, key, v)
			(*f.In)[partKey] = partVal
		}
	}

	rawSorts := readCSV(qs, "sort", v)
	slog.Debug("Read sorts", "sorts", rawSorts)
	var sort SortMatch
	for _, rawSort := range rawSorts {
		if SortRex.Find(&sort, rawSort) {
			partKey := sort.Key
			partVal := SortAsc
			if sort.Dir == "-" {
				partVal = SortDesc
			}
			f.Sort.Set(partKey, partVal)
		}
	}

	if f.Search != nil && c.Search != nil {
		v.Check(c.Search(*f.Search), "search", "parameter is invalid")
	}
	if f.Sort != nil && c.Sort != nil {
		v.Check(c.Sort(*f.Sort), "sort", "parameter is invalid")
	}
	if f.In != nil && c.In != nil {
		v.Check(c.In(*f.In), "in", "parameter is invalid")
	}

	if f.Page == nil {
		val := DefaultPage
		f.Page = &val
	} else {
		if c.Page == nil {
			c.Page = DefaultPageCheck
		}
		v.Check(c.Page(*f.Page), "page", "parameter is invalid")
	}

	if f.PageSize == nil {
		val := DefaultPageSize
		f.PageSize = &val
	} else {
		if c.PageSize == nil {
			c.PageSize = DefaultPageSizeCheck
		}
		v.Check(c.PageSize(*f.PageSize), "page_size", "parameter is invalid")
	}

	return f
}
