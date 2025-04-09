package main

import (
	"sync"

	"github.com/dusktreader/the-hunt/internal/data"
	"github.com/dusktreader/the-hunt/internal/mailer"
)

type application struct {
	config	data.Config
	models	data.Models
	mailer	*mailer.Mailer
	waiter	*sync.WaitGroup
}
