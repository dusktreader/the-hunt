package main

import (
	"net/http"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/types"
)

type healthResponse struct {
	Status		string				`json:"status"`
	Environment	types.Environment	`json:"environment"`
	Version		string				`json:"version"`
}


func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	hr := healthResponse{
		Status:			"available",
		Environment: 	app.config.APIEnv,
		Version:		types.Version,
	}

	jr := &data.JSONResponse{
		StatusCode:		http.StatusOK,
		Headers:		nil,
		Envelope:		data.Envelope{"health": hr},
	}
	app.writeJSON(w, jr)
}
