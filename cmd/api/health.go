package main

import (
	"net/http"

	"github.com/dusktreader/the-hunt/internal/data"
)

type healthResponse struct {
	Status		string `json:"status"`
	Environment	string `json:"environment"`
	Version		string `json:"version"`
}


func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	jr := &data.JSONResponse{
		StatusCode:		http.StatusOK,
		Headers:		nil,
		EnvelopeKey:	"health",
		Data:			&healthResponse{
							Status:			"available",
							Environment: 	app.config.env,
							Version:		version,
						},
	}
	app.writeJSON(w, jr)
}
