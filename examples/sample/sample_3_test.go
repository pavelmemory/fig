package sample

import (
	"fmt"

	"github.com/pavelmemory/fig"
	"github.com/pavelmemory/fig/examples/justpackage/otherrepos"
	"github.com/pavelmemory/fig/examples/justpackage/repos"
)

/*
We create our 'fig' injector with 'true' flag provided into the constructor function, it means that at time of injection
it will try to inject only fields that marked with 'fig' tag.
Also we user 'skip' configuration that allow us to explicitly mark fields that we do not want to be injected.
*/
func ExampleSimpleInjectionOnlyIfFigTagPresent() {
	injector := fig.New(true)
	FatalIfError(func() error {
		return injector.Register(
			&repos.MemUserRepo{Prefix: "1."},
		)
	})

	controller := struct {
		UserRepo_1   repos.UserRepo
		UserRepo_2   repos.UserRepo           `fig:""`
		FileUserRepo *otherrepos.FileUserRepo `fig:"skip[true]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(&controller)
	})

	if controller.UserRepo_1 == nil {
		fmt.Println("dependency without 'fig' tag configuration was not injected")
	}
	controller.UserRepo_2.Find("Pavlo")
	if controller.FileUserRepo == nil {
		fmt.Println("dependency with skip configuration was not injected")
	}
	// Output:
	// dependency without 'fig' tag configuration was not injected
	// 1. find user Pavlo in memory
	// dependency with skip configuration was not injected
}
