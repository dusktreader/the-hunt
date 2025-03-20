package data

import "time"

type Config struct {
	APIPort	int		`env:"API_PORT" envDefault:"4000"`
	APIEnv	string	`env:"API_ENV" envDefault:"development"`

	DBPort	string	`env:"DB_PORT" envDefault:"5432"`
	DBUser	string	`env:"DB_USER" envDefault:"compose-db-user"`
	DBPswd	string	`env:"DB_PSWD" envDefault:"compose-db-pswd"`
	DBName	string	`env:"DB_PSWD" envDefault:"compose-db-name"`
	DBHost	string	`env:"DB_PSWD" envDefault:"db"`

	DBMaxOpenConns	int				`env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdleConns	int				`env:"DB_MAX_IDLE_CONNS" envDefault:"25"`
	DBMaxIdleTime	time.Duration	`env:"DB_MAX_IDLE_TIME"  envDefault:"15m"`
	DBQueryTimeout	time.Duration	`env:"DB_QUERY_TIMEOUT"  envDefault:"3s"`
}
