package simple

import (
	"log"
	"os"

	"github.com/pavelmemory/fig"
	"github.com/pavelmemory/fig/example/src/otherrepos"
	"github.com/pavelmemory/fig/example/src/repos"
)

func FatalIfError(command func() error) {
	if err := command(); err != nil {
		log.Printf("%#v", err)
		os.Exit(1)
	}
}

func ExampleSimpleInjection() {
	os.Setenv("DB_URL", "some:db:connection")
	defer os.Unsetenv("DB_URL")
	injector := fig.New(false)
	FatalIfError(func() error {
		return injector.Register(
			&repos.MemUserRepo{Prefix: "1)"},
			new(otherrepos.MemUserRepo),

			otherrepos.FileUserRepo{},
			&repos.FileUserRepo{Prefix: "2)"},

			new(repos.MemOrderRepo),
		)
	})

	controller := struct {
		UserRepo repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/example/src/repos/MemUserRepo]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(&controller)
	})

	controller.UserRepo.Find("Pavlo")
	// Output:
	// 1) find user Pavlo in memory
}

