package types

type Environment string

const EnvProd Environment = "production"
const EnvStaging Environment = "staging"
const EnvQA Environment = "qa"
const EnvDev Environment = "development"

func (e Environment) IsDev() bool {
	return e != EnvProd && e != EnvStaging
}
