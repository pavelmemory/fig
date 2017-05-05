package main

import (
	"fmt"

	"github.com/pavelmemory/fig/di"
	"github.com/pavelmemory/fig/repos"
	"github.com/pavelmemory/fig/services"
	"os"
	"github.com/pavelmemory/fig/repos2"
)

func main() {
	os.Setenv("ENV_NAME", "DEV")
	fig := di.New(false)
	if err := fig.Register(
		&repos.MemUserRepo{Message: "Memory"},
		&repos2.MemUserRepo{Message: "Memory"},
		&repos2.FileUserRepo{Message: "File"},
		&repos.FileUserRepo{Message: "File"},
		repos.FileBookRepo{},
		new(repos.MemOrderRepo),
		new(repos.Module),
	); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var ovs = new(services.OracleValidationService)
	if err := fig.Initialize(ovs); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("ovs.Validate()", ovs.Validate())
	fmt.Println("ovs.Find()", ovs.Find())
	fmt.Println("ovs.(*repos.Module).Name()", ovs.Rps.Name)
	fmt.Println("ovs.(repos.UserRepo).Find()", ovs.URepo == nil)
}
