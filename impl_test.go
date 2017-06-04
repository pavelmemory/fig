package fig

import (
	"fmt"
	"os"
	"testing"

	"github.com/pavelmemory/fig/examples/justpackage/otherrepos"
	"github.com/pavelmemory/fig/examples/justpackage/repos"
)

func FatalIfError(command func() error) {
	if err := command(); err != nil {
		panic(fmt.Sprintf("%#v", err))
	}
}

func ExpectError(err error, t *testing.T, holder interface{}, expected error) {
	if err == nil {
		t.Fatalf("We expect error for: %#v", holder)
	}
	t.Log(err)
	if figError, ok := err.(FigError); ok {
		if figError.Error_ != expected {
			t.Error(fmt.Sprintf("Unexpected generic cause of error: %#v", figError))
		}
	} else {
		t.Fatal(fmt.Sprintf("Unexpected type of error: %#v", err))
	}
}

func TestFig_InitializeNil(t *testing.T) {
	injector := New(false)
	err := injector.Initialize(nil)
	ExpectError(err, t, nil, ErrorCannotBeHolder)
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
		UserRepo repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/examples/justpackage/repos/FileUserRepo]"`
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
	ExpectError(err, t, holder, ErrorCannotDecideImplementation)
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
		repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/examples/justpackage/otherrepos/MemUserRepo]"`
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
		UserRepo repos.UserRepo `fig:"impl[github.com/pavelmemory/fig/examples/justpackage/repos/MemUserRepo]"`
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
	ExpectError(err, t, nil, ErrorCannotDecideImplementation)
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

	if holder.FirstLevel == nil {
		t.Error("Nested structs were not populated properly")
	}

	if holder.FirstLevel.SecondLevel == nil {
		t.Error("Nested structs were not populated properly")
	}

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

func TestInitialize_ErrorInRecursiveInjection(t *testing.T) {
	injector := New(false)
	holder := new(holderWithReferenceFields)

	err := injector.Initialize(holder)
	ExpectError(err, t, holder, ErrorCannotDecideImplementation)
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
		t.Fatal("Nested structs were not populated properly")
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
	ExpectError(err, t, nil, ErrorCannotDecideImplementation)
}

func TestInitialize_IncorrectFigTagConfigSKIP(t *testing.T) {
	injector := New(false)
	incorrectDefinitionsOfFigTag := []interface{}{
		&struct {
			F string `fig:"skip[true"`
		}{},
		&struct {
			F string `fig:"skip[]"`
		}{},
		&struct {
			F string `fig:"skip[   ]"`
		}{},
		&struct {
			F string `fig:"skip[12]"`
		}{},
	}
	validateFigTagConfigIncorrect(incorrectDefinitionsOfFigTag, injector, t, ErrorIncorrectTagConfiguration)
}

func TestInitialize_IncorrectFigTagConfigENV(t *testing.T) {
	injector := New(false)
	incorrectDefinitionsOfFigTag := []interface{}{
		&struct {
			F string `fig:"env[true"`
		}{},
		&struct {
			F string `fig:"env[]"`
		}{},
	}
	validateFigTagConfigIncorrect(incorrectDefinitionsOfFigTag, injector, t, ErrorIncorrectTagConfiguration)
}

func TestInitialize_IncorrectFigTagConfigQUAL(t *testing.T) {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(new(repos.FileUserRepo), new(repos.MemUserRepo))
	})

	incorrectDefinitionsOfFigTag := []interface{}{
		&struct {
			repos.UserRepo `fig:"qual[anything"`
		}{},
		&struct {
			repos.UserRepo `fig:"qual[]"`
		}{},
	}
	validateFigTagConfigIncorrect(incorrectDefinitionsOfFigTag, injector, t, ErrorIncorrectTagConfiguration)

	incorrectDefinitionsOfFigTagConfValue := []interface{}{
		&struct {
			repos.UserRepo `fig:"qual[anything]"`
		}{},
	}
	validateFigTagConfigIncorrect(incorrectDefinitionsOfFigTagConfValue, injector, t, ErrorCannotDecideImplementation)
}

func TestInitialize_IncorrectFigTagConfigREG(t *testing.T) {
	injector := New(false)

	incorrectDefinitionsOfFigTag := []interface{}{
		&struct {
			F string `fig:"reg[some"`
		}{},
		&struct {
			F string `fig:"reg[]"`
		}{},
	}
	validateFigTagConfigIncorrect(incorrectDefinitionsOfFigTag, injector, t, ErrorIncorrectTagConfiguration)

	incorrectDefinitionsOfFigTagConfValue := []interface{}{
		&struct {
			F string `fig:"reg[some]"`
		}{},
		&struct {
			F string `fig:"reg[    ]"`
		}{},
	}
	validateFigTagConfigIncorrect(incorrectDefinitionsOfFigTagConfValue, injector, t, ErrorCannotDecideImplementation)
}

func TestInitialize_IncorrectFigTagConfigIMPL(t *testing.T) {
	injector := New(false)
	FatalIfError(func() error {
		return injector.Register(new(repos.FileUserRepo), new(repos.MemUserRepo))
	})

	incorrectDefinitionsOfFigTag := []interface{}{
		&struct {
			repos.UserRepo `fig:"impl[anything"`
		}{},
		&struct {
			repos.UserRepo `fig:"impl[]"`
		}{},
	}
	validateFigTagConfigIncorrect(incorrectDefinitionsOfFigTag, injector, t, ErrorIncorrectTagConfiguration)

	incorrectDefinitionsOfFigTagConfValue := []interface{}{
		&struct {
			repos.UserRepo `fig:"impl[   ]"`
		}{},
		&struct {
			repos.UserRepo `fig:"impl[12]"`
		}{},
	}
	validateFigTagConfigIncorrect(incorrectDefinitionsOfFigTagConfValue, injector, t, ErrorCannotDecideImplementation)
}

func validateFigTagConfigIncorrect(incorrectDefinitionsOfFigTag []interface{}, injector *Fig, t *testing.T, expErr error) {
	for _, incorrectDefinitionOfFigTag := range incorrectDefinitionsOfFigTag {
		err := injector.Initialize(incorrectDefinitionOfFigTag)
		ExpectError(err, t, nil, expErr)
	}
}

func TestRegister_Errors(t *testing.T) {
	injector := New(false)
	for _, cannotBeRegistered := range []interface{}{
		100,
		make(map[string]struct{}),
		nil,
		[]byte{},
	} {
		nilRegistrationError := injector.Register(cannotBeRegistered)
		if nilRegistrationError == nil {
			t.Fatal("We must not be able to register nil")
		}
		if figError, ok := nilRegistrationError.(FigError); ok {
			if figError.Error_ != ErrorCannotBeRegistered {
				t.Error(fmt.Sprintf("Unexpected generic cause of error: %#v", figError))
			}
		} else {
			t.Fatalf("Unexpected type of error: %#v", nilRegistrationError)
		}
	}
}

func TestRegisterValue_Errors(t *testing.T) {
	injector := New(false)
	nilRegistrationError := injector.RegisterValue("file_user_repo", nil)
	if nilRegistrationError == nil {
		t.Fatal("We must not be able to register nil value")
	}
	if figError, ok := nilRegistrationError.(FigError); ok {
		if figError.Error_ != ErrorCannotBeRegistered {
			t.Error(fmt.Sprintf("Unexpected generic cause of error: %#v", figError))
		}
	} else {
		t.Fatalf("Unexpected type of error: %#v", nilRegistrationError)
	}

	FatalIfError(func() error {
		return injector.RegisterValue("a", 1)
	})

	errValueOverriden := injector.RegisterValue("a", 2)
	if errValueOverriden == nil {
		t.Fatal("We need to know that we override existing value")
	}
	if figError, ok := errValueOverriden.(FigError); ok {
		if figError.Error_ != ErrorRegisteredValueOverridden {
			t.Error(fmt.Sprintf("Unexpected generic cause of error: %#v", figError))
		}
	} else {
		t.Fatalf("Unexpected type of error: %#v", errValueOverriden)
	}
}

func TestRegisterValues_Errors(t *testing.T) {
	injector := New(false)
	nilRegistrationError := injector.RegisterValues(map[string]interface{}{
		"valid":   12,
		"invalid": nil,
	})
	if nilRegistrationError == nil {
		t.Fatal("We must not be able to register nil value")
	}
	if figError, ok := nilRegistrationError.(FigError); ok {
		if figError.Error_ != ErrorCannotBeRegistered {
			t.Error(fmt.Sprintf("Unexpected generic cause of error: %#v", figError))
		}
	} else {
		t.Fatalf("Unexpected type of error: %#v", nilRegistrationError)
	}

	key1 := "a"
	key2 := "a"
	errValueOverridden := injector.RegisterValues(map[string]interface{}{
		key1: 1,
		key2: 2,
	})

	if errValueOverridden != nil {
		t.Fatal("It is impossible for maps to have same keys")
	}
}

func TestInitialize_OnlyPointersToStructsAllowed(t *testing.T) {
	injector := New(false)
	for _, holder := range []interface{}{
		100,
		"str",
		'c',
		[]byte{},
		make(map[int]struct{}),
		struct{ f string }{},
	} {
		err := injector.Initialize(holder)
		ExpectError(err, t, nil, ErrorCannotBeHolder)
	}
}

type StubUserRepo struct {
	OrderRepo repos.OrderRepo
}

var _ repos.UserRepo = new(StubUserRepo)

func (*StubUserRepo) Find(name string) {}
func (*StubUserRepo) Save(name string) {}

func TestInitialize_ErrorIfInitializationToRegisteredValueFailed(t *testing.T) {
	injector := New(false)
	sur := new(StubUserRepo)
	FatalIfError(func() error {
		return injector.Register(sur)
	})

	holder := &struct {
		repos.UserRepo
	}{}
	err := injector.Initialize(holder)
	ExpectError(err, t, nil, ErrorCannotDecideImplementation)
}

type A struct{ Pb *B }
type B struct{ Pa *A }

func TestInitialize_CyclicStructReferences(t *testing.T) {
	injector := New(false)
	a := new(A)
	b := new(B)
	FatalIfError(func() error {
		return injector.Register(a, b)
	})

	holder := &struct {
		PA *A
	}{}
	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	if holder.PA != a {
		t.Error("Incorrect injection")
	}

	if holder.PA.Pb != b {
		t.Error("Incorrect injection")
	}

	if holder.PA.Pb.Pa != a {
		t.Error("Incorrect injection")
	}
}

func TestInitialize_MapField(t *testing.T) {
	injector := New(false)

	holder := &struct {
		IntToInt     map[int]int
		StringToBool map[string]bool
		AToB         map[A]B
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})
	if holder.IntToInt == nil {
		t.Error("map[int]int not initialized")
	}
	if holder.StringToBool == nil {
		t.Error("map[string]bool not initialized")
	}
	if holder.AToB == nil {
		t.Error("map[A]B not initialized")
	}
}

func TestInitialize_ChanField(t *testing.T) {
	injector := New(false)

	holder := &struct {
		InboundQueue  <-chan repos.UserRepo
		OutboundQueue chan<- repos.UserRepo `fig:"size[2]"`
		Limiter       chan struct{}
		Lock          chan struct{} `fig:"size[0]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})
	if holder.InboundQueue == nil {
		t.Error("<-chan repos.UserRepo not initialized")
	}
	if holder.OutboundQueue == nil {
		t.Error("chan<- repos.UserRepo not initialized")
	}
	if cap(holder.OutboundQueue) != 2 {
		t.Error("chan<- repos.UserRepo size is not 2 as specified in configuration")
	}
	if holder.Limiter == nil {
		t.Error("chan struct{} not initialized")
	}
	if cap(holder.Limiter) != 1 {
		t.Error("chan struct{} size is not default size of 1")
	}
	if holder.Lock == nil {
		t.Error("chan struct{} not initialized")
	}
	if cap(holder.Lock) != 0 {
		t.Error("chan struct{} size is not 0")
	}
}

func TestInitialize_ChanFieldBadSizeConfiguration(t *testing.T) {
	injector := New(false)

	holder := &struct {
		Limiter chan struct{} `fig:"size[abc]"`
	}{}

	err := injector.Initialize(holder)
	ExpectError(err, t, holder, ErrorIncorrectTagConfiguration)
}

func TestInitialize_Slice(t *testing.T) {
	injector := New(false)

	holder := &struct {
		SizeDefault       []string
		Size10            []int `fig:"size[10]"`
		Capacity10        []int `fig:"cap[10]"`
		Size10Capacity100 []int `fig:"size[10] cap[100]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(holder)
	})

	if len(holder.SizeDefault) != 0 {
		t.Error("Expected default size of 0")
	}
	if cap(holder.SizeDefault) != 0 {
		t.Error("Expected default capacity same to size")
	}
	if len(holder.Size10) != 10 {
		t.Error("Expected provided size of 10")
	}
	if cap(holder.Capacity10) != 10 {
		t.Error("Expected provided capacity of 10")
	}
	if len(holder.Size10Capacity100) != 10 {
		t.Error("Expected provided size of 10")
	}
	if cap(holder.Size10Capacity100) != 100 {
		t.Error("Expected capacity f 100")
	}
}

func TestInitialize_SliceSizeBiggerCapacity(t *testing.T) {
	injector := New(false)

	holder := &struct {
		SizeDefaultCapacity10 []int `fig:"size[11] cap[10]"`
	}{}

	err := injector.Initialize(holder)
	ExpectError(err, t, holder, ErrorIncorrectTagConfiguration)
}

func TestInitialize_SliceWrongConfiguration(t *testing.T) {
	injector := New(false)

	holders := []interface{}{
		&struct {
			WrongSize []int `fig:"size[!]"`
		}{},
		&struct {
			WrongCapacity []float32 `fig:"cap[arf]"`
		}{},
	}
	for _, holder := range holders {
		err := injector.Initialize(holder)
		ExpectError(err, t, holder, ErrorIncorrectTagConfiguration)
	}
}

func TestInitialize_FunctionsNotSupported(t *testing.T) {
	injector := New(false)

	holder := &struct {
		Func func()
	}{}
	err := injector.Initialize(holder)
	ExpectError(err, t, holder, ErrorCannotBeHolder)
}
