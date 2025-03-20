package data

import (
	"database/sql"
	"time"
)

type ModelConfig struct {
	QueryTimeout time.Duration
}

func NewModelConfig(cfg ...Config) ModelConfig {
	mcfg := ModelConfig{}
	if len(cfg) == 1 {
		mcfg = ModelConfig{QueryTimeout: cfg[0].DBQueryTimeout}
	}
	return mcfg
}

type Models struct {
	Company CompanyModel
}

func NewModels(db *sql.DB, cfg ModelConfig) Models {
	return Models{
		Company: CompanyModel{DB: db, CFG: cfg},
	}
}
