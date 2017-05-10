package main

import (
	"github.com/pavelmemory/fig/fig"
	"github.com/pavelmemory/fig/repos"
	"github.com/pavelmemory/fig/repos2"
	"os"
	"testing"
)

func TestFig_InitializeStructWithInterfaces(t *testing.T) {
	injector := fig.New(false)
	injector.Register(
		&repos.FileUserRepo{Message: "Message"},
		&repos.FileBookRepo{},
		new(repos.MemOrderRepo),
	)

	rps := new(repos.Module)
	if err := injector.Initialize(rps); err != nil {
		t.Fatal(err)
	}

	find := rps.Find()
	if find[0] != "file" || find[1] != "Message" {
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
	injector := fig.New(false)
	injector.Register(
		&repos.FileUserRepo{Message: "File"},
		&repos.MemUserRepo{Message: "Mem"},
		&repos.FileBookRepo{},
		new(repos.MemOrderRepo),
	)

	rps := new(repos.Module)
	if err := injector.Initialize(rps); err != nil {
		t.Error(err)
	}

	find := rps.Find()
	if find[0] != "file" || find[1] != "File" {
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
	injector := fig.New(false)
	injector.Register(
		&repos.FileUserRepo{Message: "Message"},
		&repos.FileBookRepo{},
		repos.MemOrderRepo{},
	)

	repo := struct {
		repos.UserRepo
		*repos.FileBookRepo
		repos.MemOrderRepo
	}{}
	if err := injector.Initialize(&repo); err != nil {
		t.Fatal(err)
	}

	find := repo.Find()
	if find[0] != "file" || find[1] != "Message" {
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
	injector := fig.New(false)
	injector.Register(
		&repos.FileUserRepo{Message: "Message"},
		&repos2.FileUserRepo{Message: "Message2"},
	)

	repo := struct {
		repos.UserRepo
	}{}
	err := injector.Initialize(&repo)
	if err == nil {
		t.Error("Multiple implementations registered require implicit definition of one to choose")
	}
	figErr := err.(fig.FigError)
	if figErr.Error_ != fig.ErrorCannotDecideImplementation {
		t.Log(err)
		t.Error("unexpected error")
	}
}

func TestFig_Initialize_MultipleImplsWithSameStructNameWithExplicitDefinition(t *testing.T) {
	injector := fig.New(false)
	injector.Register(
		&repos.FileUserRepo{Message: "Message"},
		&repos2.FileUserRepo{Message: "Message2"},
	)

	repo := struct {
		repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/repos2/FileUserRepo]"`
	}{}
	err := injector.Initialize(&repo)
	if err != nil {
		t.Error("Multiple implementations registered require implicit definition of one to choose")
	}
	if repo.Find()[0] != "repos2" {
		t.Errorf("Incorrect implementation injected: %s", repo.Find())
	}
}

func TestFig_Initialize_InnerFieldsShouldBeInjectedAutomaticallyIfRegistered(t *testing.T) {
	injector := fig.New(false)
	if err := injector.Register(
		&repos.MemUserRepo{Message: "Memory"},
		&repos.FileUserRepo{Message: "File"},
		repos.FileBookRepo{},
		new(repos.MemOrderRepo),
	); err != nil {
		t.Error(err)
	}

	nested := struct {
		*repos.Module
	}{}
	if err := injector.Initialize(&nested); err != nil {
		t.Fatal(err)
	}
	if nested.Module.BookRepo == nil {
		t.Error("Not initialized")
	}
}

func Test_Initialize_EnvVar(t *testing.T) {
	envKey, envValue := "ENV_NAME", "DEV"
	os.Setenv(envKey, envValue)
	defer func() {
		os.Unsetenv(envKey)
	}()
	injector := fig.New(false)

	holder := struct {
		EnvName string `fig:"env[ENV_NAME]"`
	}{}

	if err := injector.Initialize(&holder); err != nil {
		t.Error(err)
	}

	if holder.EnvName != envValue {
		t.Error("Env var was not set")
	}
}

func Test_Initialize_Skip(t *testing.T) {
	envKey, envValue := "ENV_NAME", "DEV"
	os.Setenv(envKey, envValue)
	defer func() {
		os.Unsetenv(envKey)
	}()
	injector := fig.New(false)
	injector.Register(&repos2.FileUserRepo{})

	holder := struct {
		UserRepoShouldBeNil  repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/repos2/FileUserRepo] skip[true]"`
		UserRepoShouldBeInit repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/repos2/FileUserRepo] skip[false]"`
		EnvName              string         `fig:"env[ENV_NAME] skip[true]"`
	}{}

	if err := injector.Initialize(&holder); err != nil {
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

func Test_Initialize_RegisterValue(t *testing.T) {
	injector := fig.New(false)
	regValue := new(int)
	*regValue = 100
	injector.RegisterValue("regKey", regValue)
	injector.RegisterValue("regKey2", rune(100))

	holder := struct {
		RegValue     *int `fig:"reg[regKey]"`
		RegValueSkip rune `fig:"reg[regKey2] skip[true]"`
	}{}

	if err := injector.Initialize(&holder); err != nil {
		t.Fatal(err)
	}

	if *holder.RegValue != 100 {
		t.Error("RegValue should be set")
	}

	if holder.RegValueSkip != rune(0) {
		t.Error("RegValue should not be set")
	}
}

func Test_Initialize_RegisterValueNotFound(t *testing.T) {
	injector := fig.New(false)

	holder := struct {
		RegValue int `fig:"reg[regKey]"`
	}{}

	err := injector.Initialize(&holder)
	if err == nil {
		t.Fatal("Should singal error")
	}
	figErr := err.(fig.FigError)
	if figErr.Error_ != fig.ErrorCannotDecideImplementation {
		t.Error("Should be that type of error")
	}
}

func Test_Initialize_OnlyWithFigTag(t *testing.T) {
	injector := fig.New(true)
	injector.Register(new(repos.FileUserRepo))
	holder := struct {
		DoesNotNeedToInject repos.UserRepo
		NeedToInject        repos.UserRepo `fig:""`
	}{}

	err := injector.Initialize(&holder)
	if err != nil {
		t.Fatal(err)
	}
	if holder.DoesNotNeedToInject != nil {
		t.Error("Fields without `injector` tag should not be injected")
	}
	if holder.NeedToInject == nil {
		t.Error("Fields without `injector` must be injected")
	}
}

func Test_Initialize_RecursiveInjectionToUnnamedStructs(t *testing.T) {
	injector := fig.New(false)
	injector.Register(new(repos.FileUserRepo))
	holder := struct {
		FirstLevel struct {
			JustSimpleFieldOnFirstLevel string
			SecondLevel                 struct {
				NeedToBeInjected             repos.UserRepo
				JustSimpleFieldOnSecondLevel string
			}
		}
	}{}

	err := injector.Initialize(&holder)
	if err != nil {
		t.Fatal(err)
	}
	if holder.FirstLevel.SecondLevel.NeedToBeInjected == nil {
		t.Error("Nested structs were not populated properly")
	}
}

type secondLevelReferenceStruct struct {
	NeedToBeInjected             repos.UserRepo
	JustSimpleFieldOnSecondLevel string `fig:"env[ENV_NAME]"`
}

type firstLevelReferenceStruct struct {
	JustSimpleFieldOnFirstLevel string
	SecondLevel                 *secondLevelReferenceStruct
}

type holderWithReferenceFields struct {
	FirstLevel *firstLevelReferenceStruct
}

func Test_Initialize_RecursiveInjectionToReferenceFields(t *testing.T) {
	envKey, envValue := "ENV_NAME", "DEV"
	os.Setenv(envKey, envValue)
	defer func() {
		os.Unsetenv(envKey)
	}()

	injector := fig.New(false)
	injector.Register(&repos.FileUserRepo{Message: "File"})
	holder := holderWithReferenceFields{}

	err := injector.Initialize(&holder)
	if err != nil {
		t.Fatal(err)
	}
	if holder.FirstLevel.SecondLevel.NeedToBeInjected == nil {
		t.Error("Nested structs were not populated properly")
	}
	if holder.FirstLevel.SecondLevel.NeedToBeInjected.Find()[1] != "File" {
		t.Error("Incorrect implementation was injected")
	}
	if holder.FirstLevel.SecondLevel.JustSimpleFieldOnSecondLevel != "DEV" {
		t.Error("Simple value from env was not injected")
	}
}

type firstLevelReferenceStructWithoutDependencies struct {
	SecondLevel *secondLevelReferenceStructWithoutDependencies
}

type secondLevelReferenceStructWithoutDependencies struct {
	JustSimpleFieldOnSecondLevel string
}

func Test_Initialize_RecursiveInjectionToReferenceFieldsWithoutDependencies(t *testing.T) {
	injector := fig.New(false)
	holder := struct {
		FirstLevel *firstLevelReferenceStructWithoutDependencies
	}{}

	err := injector.Initialize(&holder)
	if err != nil {
		t.Fatal(err)
	}
	if holder.FirstLevel.SecondLevel == nil {
		t.Error("Nested structs were not populated properly")
	}
}

func Test_Initialize_PointerToPointer(t *testing.T) {
	injector := fig.New(false)
	if err := injector.Register(new(repos.FileBookRepo)); err != nil {
		t.Fatal(err)
	}

	holder := struct {
		RefToRef ****repos.FileBookRepo
	}{}

	if err := injector.Initialize(&holder); err != nil {
		t.Fatal(err)
	}

	if holder.RefToRef == nil {
		t.Error("Nested structs were not populated properly")
	}
	if (***holder.RefToRef).Get() != "qwqe" {
		t.Error("Incorrect impl injected")
	}
}

type StringGetter interface {
	Get() string
}

type StringGetterWithQualifier struct {
	qualifier string
}

func (sgwq *StringGetterWithQualifier) Get() string {
	return sgwq.qualifier
}

func (sgwq *StringGetterWithQualifier) Qualify() string {
	return sgwq.qualifier
}

type StringGetterWithoutQualifier struct {
}

func (sgwq *StringGetterWithoutQualifier) Get() string {
	return "something"
}

func Test_Initialize_QualifyDefinedCorrectly(t *testing.T) {
	injector := fig.New(false)
	if err := injector.Register(
		&StringGetterWithQualifier{},
		&StringGetterWithQualifier{qualifier: "ass"},
		new(StringGetterWithoutQualifier),
		&StringGetterWithQualifier{qualifier: "tits"},
	); err != nil {
		t.Fatal(err)
	}

	holder := struct {
		ByQualifier StringGetter `fig:"qual[tits]"`
		ByImpl      StringGetter `fig:"impl[github.com/pavelmemory/fig/StringGetterWithoutQualifier]"`
	}{}
	if err := injector.Initialize(&holder); err != nil {
		t.Fatal(err)
	}
	if holder.ByQualifier == nil {
		t.Fatal("implementation was not injeted when qualifier in use")
	}
	if holder.ByQualifier.Get() != "tits" {
		t.Error("incorrect implementation was injeted with qualifier in use")
	}
	if holder.ByImpl == nil {
		t.Fatal("implementation was not injeted when impl defined explicitly")
	}
	if holder.ByImpl.Get() != "something" {
		t.Error("incorrect implementation was injeted when impl defined explicitly")
	}
}

func Test_Initialize_QualifyNotDefinedButTagProvided(t *testing.T) {
	injector := fig.New(false)
	if err := injector.Register(
		new(StringGetterWithQualifier),
		new(StringGetterWithoutQualifier),
	); err != nil {
		t.Fatal(err)
	}

	holder := struct {
		ByQualifier StringGetter `fig:"qual[tits]"`
	}{}
	err := injector.Initialize(&holder)
	if err == nil {
		t.Fatal("there must be error because nothing in two registered equals to 'qual' value")
	}
	figErr := err.(fig.FigError)
	if figErr.Error_ != fig.ErrorCannotDecideImplementation {
		t.Log(err)
		t.Error("unexpected error")
	}
}
