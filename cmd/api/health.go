package main

import (
	"net/http"

	"github.com/dusktreader/the-hunt/internal/data"
)

type healthResponse struct {
	Status		string			`json:"status"`
	Environment	string			`json:"environment"`
	Version		string			`json:"version"`
	Config		*data.Config	`json:"config"`
}


func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	hr := healthResponse{
		Status:			"available",
		Environment: 	app.config.APIEnv,
		Version:		version,
	}
	if hr.Environment != "production" {
		hr.Config = &app.config
	}

	jr := &data.JSONResponse{
		StatusCode:		http.StatusOK,
		Headers:		nil,
		Envelope:		data.Envelope{"health": hr},
	}
	app.writeJSON(w, jr)
}
