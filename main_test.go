package main

import (
	"github.com/pavelmemory/fig/di"
	"github.com/pavelmemory/fig/repos"
	"os"
	"testing"
	"github.com/pavelmemory/fig/repos2"
	"fmt"
)

func TestFig_InitializeStructWithInterfaces(t *testing.T) {
	fig := di.New()
	fig.Register(
		&repos.FileUserRepo{Message: "Message"},
		&repos.FileBookRepo{},
		new(repos.MemOrderRepo),
	)

	rps := new(repos.Module)
	if err := fig.Initialize(rps); err != nil {
		t.Error(err)
	}

	find := rps.Find()
	if find[0] != "repos" || find[1] != "Message" {
		t.Error("Find() failed")
	}

	if rps.Get() != "qwqe" {
		t.Error("Get() failed")
	}

	if !rps.Make() {
		t.Error("Make() failed")
	}
}

func TestFig_InitializeStructWithInterfacesMultipleImpls(t *testing.T) {
	fig := di.New()
	fig.Register(
		&repos.FileUserRepo{Message: "File"},
		&repos.MemUserRepo{Message: "Mem"},
		&repos.FileBookRepo{},
		new(repos.MemOrderRepo),
	)

	rps := new(repos.Module)
	if err := fig.Initialize(rps); err != nil {
		t.Error(err)
	}

	find := rps.Find()
	if find[0] != "repos" || find[1] != "File" {
		t.Error("Find() failed")
	}

	if rps.Get() != "qwqe" {
		t.Error("Get() failed")
	}

	if !rps.Make() {
		t.Error("Make() failed")
	}
}

func TestFig_Initialize_UnnamedStructWithInterfacesAndValue(t *testing.T) {
	fig := di.New()
	fig.Register(
		&repos.FileUserRepo{Message: "Message"},
		&repos.FileBookRepo{},
		repos.MemOrderRepo{},
	)

	repo := struct {
		repos.UserRepo
		*repos.FileBookRepo
		repos.MemOrderRepo
	}{}
	if err := fig.Initialize(&repo); err != nil {
		t.Fatal(err)
	}

	find := repo.Find()
	if find[0] != "repos" || find[1] != "Message" {
		t.Error("Find() failed")
	}

	if repo.Get() != "qwqe" {
		t.Error("Get() failed")
	}

	if !repo.Make() {
		t.Error("Make() failed")
	}
}

func TestFig_Initialize_MultipleImplsWithSameStructNameWithoutExplicitDefinition(t *testing.T) {
	fig := di.New()
	fig.Register(
		&repos.FileUserRepo{Message: "Message"},
		&repos2.FileUserRepo{Message: "Message2"},
	)

	repo := struct {
		repos.UserRepo
	}{}
	err := fig.Initialize(&repo)
	if err == nil {
		t.Error("Multiple implementations registered require implicit definition of one to choose")
	}
	fmt.Println(err)
}

func TestFig_Initialize_MultipleImplsWithSameStructNameWithExplicitDefinition(t *testing.T) {
	fig := di.New()
	fig.Register(
		&repos.FileUserRepo{Message: "Message"},
		&repos2.FileUserRepo{Message: "Message2"},
	)

	repo := struct {
		repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/repos2/FileUserRepo]"`
	}{}
	err := fig.Initialize(&repo)
	if err != nil {
		t.Error("Multiple implementations registered require implicit definition of one to choose")
	}
	fmt.Println(repo.Find())
	if repo.Find()[0] != "repos2" {
		t.Errorf("Incorrect implementation injected: %s", repo.Find())
	}
}

func TestFig_Initialize_InnerFieldsShouldPopulateAutomatically(t *testing.T) {
	fig := di.New()
	if err := fig.Register(
		&repos.MemUserRepo{Message: "Memory"},
		&repos.FileUserRepo{Message: "File"},
		repos.FileBookRepo{},
		new(repos.MemOrderRepo),
		new(repos.Module),
	); err != nil {
		t.Error(err)
	}

	nested := struct {
		*repos.Module
	}{}
	if err := fig.Initialize(&nested); err != nil {
		fmt.Println(err.Error())
		t.Fatal(err)
	}

	if nested.Module.BookRepo == nil {
		t.Error("Not initialized")
	}
}

func Test_Initialize_EnvVar(t *testing.T) {
	envValue := "DEV"
	os.Setenv("ENV_NAME", envValue)
	fig := di.New()

	holder := struct {
		EnvName string `fig:"env[ENV_NAME]"`
	}{}

	if err := fig.Initialize(&holder); err != nil {
		t.Error(err)
	}

	if holder.EnvName != envValue {
		t.Error("Env var was not set")
	}
	os.Unsetenv("ENV_NAME")
}


func Test_Initialize_Skip(t *testing.T) {
	envValue := "DEV"
	os.Setenv("ENV_NAME", envValue)
	fig := di.New()
	fig.Register(&repos2.FileUserRepo{})

	holder := struct {
		UserRepoShouldBeNil repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/repos2/FileUserRepo] skip[true]"`
		UserRepoShouldBeInit repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/repos2/FileUserRepo] skip[false]"`
		EnvName string `fig:"env[ENV_NAME] skip[true]"`
	}{}

	if err := fig.Initialize(&holder); err != nil {
		t.Fatal(err)
	}

	if holder.EnvName != "" {
		t.Error("Env var should not be set")
	}
	if holder.UserRepoShouldBeNil != nil {
		t.Error("Should not be set because skip = true provided")
	}
	if holder.UserRepoShouldBeInit == nil {
		t.Error("Should be set because skip = false provided")
	}
	os.Unsetenv("ENV_NAME")
}
