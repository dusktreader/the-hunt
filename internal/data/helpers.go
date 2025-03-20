package data

import (
	"net/http"
)

type JSONResponse struct {
	Data		any
	StatusCode	int
	Headers		http.Header
	EnvelopeKey	string
}

type Envelope map[string]any

type ErrorPackage struct {
	Error		error	`json:"-"`
	Message		string	`json:"message"`
	LogMessage	string	`json:"-"`
	Details		any		`json:"details,omitempty"`
	StatusCode	int		`json:"-"`
}
