package data

import "slices"

type SearchMap map[string]string
type InMap map[string]string

type Filters struct {
	Search		*SearchMap
	Sort		*string
	In			*InMap
	SortAsc		*bool
	Page		*int
	PageSize	*int
}

type IntRange struct {
	Min int
	Max int
}

type FilterConstraints struct {
	Search		func(SearchMap) bool
	Sort		func(string) bool
	In			func(InMap) bool
	Page		func(int) bool
	PageSize	func(int) bool
}

type SearchFields []string
type SortFields []string
type InFields []string

func (sf SearchFields) Check(sm SearchMap) bool {
	for k := range sm {
		if !slices.Contains(sf, k) {
			return false
		}
	}
	return true
}

func (sf SortFields) Check(s string) bool {
	return slices.Contains(sf, s)
}

func (nf InFields) Check(nm InMap) bool {
	for k := range nm {
		if !slices.Contains(nf, k) {
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
