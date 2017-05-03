package services

import "github.com/pavelmemory/fig/repos"

type ValidationService interface {
	Validate() bool
	Find() []string
}

type OracleValidationService struct {
	Rps   *repos.Module
	URepo repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/repos/FileUserRepo] env[JERONIMO]     skip[true]   "`
}

func (ovs *OracleValidationService) Validate() bool {
	return ovs.Rps.Make()
}

func (ovs *OracleValidationService) Find() []string {
	return ovs.Rps.Find()
}
