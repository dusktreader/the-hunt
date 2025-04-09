package data

import (
	"time"

	"golang.org/x/time/rate"

	"github.com/dusktreader/the-hunt/internal/types"
)

type Config struct {
	APIPort	int					`env:"API_PORT" envDefault:"4000"`
	APIEnv	types.Environment	`env:"API_ENV"  envDefault:"development"`

	DBPort	string	`env:"DB_PORT" envDefault:"5432"`
	DBUser	string	`env:"DB_USER" envDefault:"compose-db-user"`
	DBPswd	string	`env:"DB_PSWD" envDefault:"compose-db-pswd" json:"-"`
	DBName	string	`env:"DB_NAME" envDefault:"compose-db-name"`
	DBHost	string	`env:"DB_HOST" envDefault:"db"`

	DBMaxOpenConns	int				`env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdleConns	int				`env:"DB_MAX_IDLE_CONNS" envDefault:"25"`
	DBMaxIdleTime	time.Duration	`env:"DB_MAX_IDLE_TIME"  envDefault:"15m"`
	DBQueryTimeout	time.Duration	`env:"DB_QUERY_TIMEOUT"  envDefault:"3s"`

	MailHost	string  `env:"MAIL_HOST" envDefault:"maildev"`
	MailPort	int		`env:"MAIL_PORT" envDefault:"25"`
	MailUser	string	`env:"MAIL_USER" envDefault:"compose-mail-user"`
	MailPswd	string	`env:"MAIL_PSWD" envDefault:"compose-mail-pswd" json:"-"`
	MailSndr	string	`env:"MAIL_SNDR" envDefault:"dev@localhost"`

	MailSendRetries	int				`env:"MAIL_SEND_RETRIES" envDefault:"3"`
	MailRetryDelay	time.Duration	`env:"MAIL_RETRY_DELAY" envDefault:"500ms"`

	LimitEnabled	bool			`env:"LIMIT_ENABLED" envDefault:"true"`
	LimitRPS		rate.Limit		`env:"LIMIT_RPS"     envDefault:"5.0"`
	LimitBurst		int				`env:"LIMIT_BURST"   envDefault:"10"`

	ClientCleanupInterval	time.Duration	`env:"CLIENT_CLEANUP_INTERVAL" envDefault:"1m"`
	ClientCleanupTimeout	time.Duration	`env:"CLIENT_CLEANUP_TIMEOUT"  envDefault:"3m"`

	ActivationTTL	time.Duration `env:"ACTIVATION_TTL" envDefault:"72h"`
	AuthTTL			time.Duration `env:"AUTH_TTL" envDefault:"1h"`

	AdminEmail		types.Email		`env:"ADMIN_EMAIL"`
	AdminPassword	types.PlainPW	`env:"ADMIN_PASSWORD" json:"-"`

	CORSTrustOrigins	[]string	`env:"CORS_TRUST_ORIGINS"`
}
