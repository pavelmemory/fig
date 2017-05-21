package fig

import (
	"fmt"
	"os"
	"testing"

	"github.com/pavelmemory/fig/examples/src/otherrepos"
	"github.com/pavelmemory/fig/examples/src/repos"
)

func FatalIfError(command func() error) {
	if err := command(); err != nil {
		panic(fmt.Sprintf("%#v", err))
	}
}

func TestFig_InitializeNil(t *testing.T) {
	injector := New(false)
	err := injector.Initialize(nil)
	if err == nil {
		t.Fatal("The error is mandatry in such a case")
	}

	figErr := err.(FigError)
	if figErr.Error_ != ErrorCannotBeHolder {
		t.Error("invalid error reason")
	}
}

func Example_InitializeStructWithInterfaces() {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(
			&repos.FileUserRepo{Prefix: "1."},
			new(repos.MemOrderRepo),
		)
	})

	holder := &struct {
		repos.UserRepo
		repos.OrderRepo
		ExplicitUserRepo  repos.UserRepo
		ExplicitOrderRepo repos.OrderRepo
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	holder.Find("Ivan")
	holder.Create()
	holder.ExplicitUserRepo.Find("Sasha")
	holder.ExplicitOrderRepo.Remove("2")
	// Output:
	// 1. find user Ivan in file
	// create order in memory
	// 1. find user Sasha in file
	// remove 2 order from memory
}

func Example_InitializeStructWithMultipleInterfaceImplementations() {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(
			&repos.FileUserRepo{Prefix: "1."},
			&repos.MemUserRepo{Prefix: "1."},
		)
	})

	holder := &struct {
		UserRepo repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/examples/src/repos/FileUserRepo]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	holder.UserRepo.Save("Pavlo")
	// Output:
	// 1. save user Pavlo to file
}

func TestInitializeStructWithValue(t *testing.T) {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(
			repos.MemOrderRepo{Count: 100},
		)
	})

	holder := &struct {
		MemOrderRepo repos.MemOrderRepo
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	if holder.MemOrderRepo.Count != 100 {
		t.Error("Initialization value was not injected")
	}
}

func Example_InitializeStructWithReference() {
	injector := New(false)

	holder := &struct {
		MemOrderRepo *repos.MemOrderRepo
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	holder.MemOrderRepo.Create()
	// Output:
	// create order in memory
}

func TestInitialize_InterfaceWithMultipleImplementationsWithSameStructNameWithoutExplicitDefinition(t *testing.T) {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(
			new(repos.MemUserRepo),
			new(otherrepos.MemUserRepo),
		)
	})

	holder := &struct {
		repos.UserRepo
	}{}
	err := injector.Initialize(holder)
	if err == nil {
		t.Error("Multiple implementations registered require implicit definition of one to choose")
	}
	figErr := err.(FigError)
	if figErr.Error_ != ErrorCannotDecideImplementation {
		t.Log(err)
		t.Error("unexpected error")
	}
}

func Example_InitializeInterfaceWithMultipleImplementationsWithSameStructNameWithExplicitDefinition() {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(
			new(repos.MemUserRepo),
			new(otherrepos.MemUserRepo),
		)
	})

	holder := &struct {
		repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/examples/src/otherrepos/MemUserRepo]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})
	holder.UserRepo.Save("Ivan")
	// Output:
	// save user IVAN to memory
}

func Example_InitializeInnerFieldsShouldBeInjectedAutomaticallyIfRegistered() {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(
			&repos.MemUserRepo{Prefix: "1."},
			new(repos.MemOrderRepo),
		)
	})
	holder := &struct {
		repos.UserRepo
		InnerHolder struct {
			repos.OrderRepo
		}
	}{}
	FatalIfError(func() error {
		return injector.Initialize(holder)
	})
	holder.UserRepo.Save("Julia")
	holder.InnerHolder.OrderRepo.Remove("100")
	// Output:
	// 1. save user Julia to memory
	// remove 100 order from memory
}

func TestInitialize_EnvVar(t *testing.T) {
	envKey, envValue := "ENV_NAME", "DEV"
	os.Setenv(envKey, envValue)
	defer os.Unsetenv(envKey)

	holder := &struct {
		EnvName string `fig:"env[ENV_NAME]"`
	}{}

	FatalIfError(func() error {
		return New(false).Initialize(holder)
	})

	if holder.EnvName != envValue {
		t.Error("Env var was not set")
	}
}

func Example_InitializeExplicitImplementationSpecificationSkippedIfSingleImplementationRegistered() {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(&otherrepos.FileUserRepo{})
	})

	holder := &struct {
		UserRepo repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/examples/src/repos/MemUserRepo]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})
	holder.UserRepo.Find("Lolly")
	// Output:
	// find user LOLLY in file
}

func TestInitialize_Skip(t *testing.T) {
	envKey, envValue := "ENV_NAME", "DEV"
	os.Setenv(envKey, envValue)
	defer os.Unsetenv(envKey)

	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(&otherrepos.FileUserRepo{})
	})

	holder := &struct {
		UserRepoShouldBeNil  repos.UserRepo `fig:"impl[we_dont_care] skip[true]"`
		UserRepoShouldBeInit repos.UserRepo
		DefaultString        string `fig:"env[ENV_NAME] skip[true]"`
		InitializedString    string `fig:"env[ENV_NAME] skip[false]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	if holder.UserRepoShouldBeNil != nil {
		t.Error("Should not be set because skip = true provided")
	}
	if holder.UserRepoShouldBeInit == nil {
		t.Error("Should be set because by default skip = false")
	}
	if holder.DefaultString != "" {
		t.Error("Env var must not be set")
	}
	if holder.InitializedString != envValue {
		t.Error("Env var must be set")
	}
}

func TestInitialize_RegisterValue(t *testing.T) {
	injector := New(false)
	regValue := new(int)
	*regValue = 100
	FatalIfError(func() error {
		return injector.RegisterValues(map[string]interface{}{
			"regKey":  regValue,
			"regKey2": rune(100),
			"ch1":     make(map[int]int),
		})
	})

	holder := &struct {
		RegValue          *int        `fig:"reg[regKey]"`
		RegValueSkipTrue  rune        `fig:"reg[regKey2] skip[true]"`
		RegValueSkipFalse map[int]int `fig:"reg[ch1] skip[false]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})
	if *holder.RegValue != 100 {
		t.Error("RegValue should be set")
	}

	if holder.RegValueSkipTrue != rune(0) {
		t.Error("RegValue should not be set")
	}
	holder.RegValueSkipFalse[1] = 1 // if map is not initialized panic will happen
}

func TestInitialize_RegisterValueNotFound(t *testing.T) {
	injector := New(false)

	holder := &struct {
		RegValue int `fig:"reg[regKey]"`
	}{}

	err := injector.Initialize(holder)
	if err == nil {
		t.Fatal("Should singal error")
	}
	figErr := err.(FigError)
	if figErr.Error_ != ErrorCannotDecideImplementation {
		t.Error("Should be that type of error")
	}
}

func Example_InitializeOnlyWithFigTag() {
	injector := New(true)

	FatalIfError(func() error {
		return injector.Register(new(repos.FileUserRepo))
	})

	holder := &struct {
		DoesNotNeedToInject repos.UserRepo
		NeedToInject        repos.UserRepo `fig:""`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})
	if holder.DoesNotNeedToInject != nil {
		panic("Fields without `fig` tag should not be injected")
	}
	holder.NeedToInject.Save("Igor")
	// Output:
	// save user Igor to file
}

func TestInitialize_RecursiveInjectionToUnnamedStructs(t *testing.T) {
	regValue := "DEV"
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(new(repos.FileUserRepo))
	})
	FatalIfError(func() error {
		return injector.RegisterValue("key", regValue)
	})

	holder := &struct {
		FirstLevel struct {
			JustSimpleFieldOnFirstLevel string `fig:"reg[key]"`
			SecondLevel                 struct {
				NeedToBeInjected             repos.UserRepo
				JustSimpleFieldOnSecondLevel string
			}
		}
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	if holder.FirstLevel.JustSimpleFieldOnFirstLevel != regValue {
		t.Error("Field was not injected")
	}

	if holder.FirstLevel.SecondLevel.NeedToBeInjected == nil {
		t.Error("Nested structs were not populated properly")
	}
}

type secondLevelReferenceStruct struct {
	UserRepo                     repos.UserRepo
	JustSimpleFieldOnSecondLevel string `fig:"env[ENV_NAME]"`
}

type firstLevelReferenceStruct struct {
	JustSimpleFieldOnFirstLevel string
	SecondLevel                 *secondLevelReferenceStruct
}

type holderWithReferenceFields struct {
	FirstLevel *firstLevelReferenceStruct
}

func TestInitialize_RecursiveInjectionToReferenceFields(t *testing.T) {
	envKey, envValue := "ENV_NAME", "DEV"
	os.Setenv(envKey, envValue)
	defer os.Unsetenv(envKey)

	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(new(repos.FileUserRepo))
	})
	holder := new(holderWithReferenceFields)

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	if holder.FirstLevel.SecondLevel.UserRepo == nil {
		t.Error("Nested structs were not populated properly")
	}
	if holder.FirstLevel.SecondLevel.JustSimpleFieldOnSecondLevel != "DEV" {
		t.Error("Simple value from env was not injected")
	}
}

func TestInitialize_NoRecursiveInjectionToReferenceFieldsWithoutExplicitFigTag(t *testing.T) {
	envKey, envValue := "ENV_NAME", "DEV"
	os.Setenv(envKey, envValue)
	defer os.Unsetenv(envKey)

	injector := New(true)
	FatalIfError(func() error {
		return injector.Register(new(repos.FileUserRepo))
	})

	holder := new(holderWithReferenceFields)

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	if holder.FirstLevel != nil {
		t.Error("Nested structs were not populated properly")
	}
}

type firstLevelReferenceStructWithoutDependencies struct {
	SecondLevel *secondLevelReferenceStructWithoutDependencies
}

type secondLevelReferenceStructWithoutDependencies struct {
	JustSimpleFieldOnSecondLevel string
}

func TestInitialize_RecursiveInjectionToReferenceFieldsWithoutDependencies(t *testing.T) {
	injector := New(false)
	holder := &struct {
		FirstLevel *firstLevelReferenceStructWithoutDependencies
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	if holder.FirstLevel.SecondLevel == nil {
		t.Error("Nested structs were not populated properly")
	}
}

func TestInitialize_PointerToPointer(t *testing.T) {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(new(repos.FileUserRepo))
	})

	holder := &struct {
		RefToRef ****repos.FileUserRepo
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	if holder.RefToRef == nil {
		t.Error("Nested structs were not populated properly")
	}
	(***holder.RefToRef).Save("Eva")
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

func TestInitialize_QualifyDefinedCorrectly(t *testing.T) {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(
			&StringGetterWithQualifier{},
			&StringGetterWithQualifier{qualifier: "ass"},
			new(StringGetterWithoutQualifier),
			&StringGetterWithQualifier{qualifier: "tits"},
		)
	})

	holder := &struct {
		ByQualifier StringGetter `fig:"qual[tits]"`
		ByImpl      StringGetter `fig:"impl[github.com/pavelmemory/fig/StringGetterWithoutQualifier]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

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

func TestInitialize_QualifyNotDefinedButTagProvided(t *testing.T) {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(
			new(StringGetterWithQualifier),
			new(StringGetterWithoutQualifier),
		)
	})

	holder := &struct {
		ByQualifier StringGetter `fig:"qual[tits]"`
	}{}

	err := injector.Initialize(holder)
	if err == nil {
		t.Fatal("there must be error because nothing in two registered equals to 'qual' value")
	}

	figErr := err.(FigError)
	if figErr.Error_ != ErrorCannotDecideImplementation {
		t.Log(err)
		t.Error("unexpected error")
	}
}

func TestInitialize_IncorrectFigTagConfigSKIP(t *testing.T) {
	incorrectDefinitionsOfFigTag := []interface{} {
		&struct { F string `fig:"skip"` }{},
		&struct { F string `fig:"skip[true"` }{},
		&struct { F string `fig:"skip]"` }{},
		&struct { F string `fig:"skip[]"` }{},
		&struct { F string `fig:"skip[   ]"` }{},
		&struct { F string `fig:"skip[12]"` }{},
		&struct { F string `fig:"impl[something] skip"` }{},
	}
	injector := New(false)
	for _, incorrectDefinitionOfFigTag := range incorrectDefinitionsOfFigTag {
		err := injector.Initialize(incorrectDefinitionOfFigTag)
		if err != nil {
			t.Error(fmt.Sprintf("Initialization error expected for type: %#v", incorrectDefinitionOfFigTag))
		}
	}
}