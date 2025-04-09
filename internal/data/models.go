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
	Company    CompanyModel
	User       UserModel
	Token      TokenModel
	Permission PermissionModel
}

func NewModels(db *sql.DB, cfg ModelConfig) Models {
	return Models{
		Company:    CompanyModel{DB: db, CFG: cfg},
		User:       UserModel{DB: db, CFG: cfg},
		Token:      TokenModel{DB: db, CFG: cfg},
		Permission: PermissionModel{DB: db, CFG: cfg},
	}
}
