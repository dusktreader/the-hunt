package data

import (
	"net/http"
)

type JSONResponse struct {
	Envelope   Envelope
	StatusCode int
	Headers    http.Header
}

type Envelope map[string]any

type ErrorPackage struct {
	Error      error  `json:"-"`
	Message    string `json:"message"`
	LogMessage string `json:"-"`
	Details    any    `json:"details,omitempty"`
	StatusCode int    `json:"-"`
}

type ListMetadata struct {
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
	FirstPage   int `json:"first_page"`
	LastPage    int `json:"last_page"`
	RecordCount int `json:"record_count"`
}

func NewListMetadata(f Filters, recordCount int) ListMetadata {
	if recordCount == 0 {
		return ListMetadata{}
	}

	return ListMetadata{
		CurrentPage: *f.Page,
		PageSize:    *f.PageSize,
		FirstPage:   1,
		LastPage:    (recordCount + *f.PageSize - 1) / *f.PageSize,
		RecordCount: recordCount,
	}
}
