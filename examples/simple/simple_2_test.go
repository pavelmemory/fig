package simple

import (
	"log"
	"os"

	"github.com/pavelmemory/fig"
	"github.com/pavelmemory/fig/examples/src/otherrepos"
	"github.com/pavelmemory/fig/examples/src/repos"
)

func FatalIfError(command func() error) {
	if err := command(); err != nil {
		log.Printf("%#v", err)
		os.Exit(1)
	}
}

/*
Below is an example of injection into interface type field 'UserRepo' with explicit defined implementation.
This explicit definition is mandatory because we register two struct that implement the interface and
it is our responsibility to help fig tool to decide between two.

Also there is an embedding of the interface 'OrderRepo' which is implemented only by one of registered stucts - 'MemOrderRepo'.
So no explicit definition required for this field.

We create our 'fig' injector with 'false' flag provided into the constructor function it means that at time of injection
it will try to inject all fields of struct.
So because we have field 'MemUserRepo' which has reference struct type and this struct was not registered explicitly
it will be created by 'fig' and injected automatically.
*/
func ExampleSimpleInjection() {
	os.Setenv("DB_URL", "some:db:connection")
	defer os.Unsetenv("DB_URL")

	injector := fig.New(false)
	FatalIfError(func() error {
		return injector.Register(
			&repos.MemUserRepo{Prefix: "1)"},

			otherrepos.FileUserRepo{},
			&repos.FileUserRepo{Prefix: "2)"},

			new(repos.MemOrderRepo),
		)
	})

	controller := struct {
		// explicit definition of type to be injected because multiple interface imps is registered
		UserRepo repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/examples/src/repos/MemUserRepo]"`
		// Injection of embedded interfaces is also supported
		repos.OrderRepo
		// Struct
		MemUserRepo  *otherrepos.MemUserRepo
		FileUserRepo *otherrepos.FileUserRepo `fig:"skip[true]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(&controller)
	})

	controller.UserRepo.Find("Pavlo")
	controller.Create()
	controller.MemUserRepo.Save("Linda")
	// Output:
	// 1) find user Pavlo in memory
	// create order in memory
	// save user LINDA to memory
}
