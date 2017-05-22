package fig

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
)

type Qualifier interface {
	Qualify() string
}

const (
	// fig tag itself
	FIG_TAG = "fig"
	// configurations for fig tag
	IMPL_TAG_KEY = "impl"
	ENV_TAG_KEY  = "env"
	SKIP_TAG_KEY = "skip"
	REG_TAG_KEY  = "reg"
	QUAL_TAG_KEY = "qual"
)

func getConfigValueForKey(conf string, key string) (string, bool, error) {
	confHolder := key + "["
	keyStart := strings.Index(conf, confHolder)
	if keyStart < 0 {
		return "", false, nil
	}

	//if keyStart > 0 && !(conf[keyStart-1] == ' ' || conf[keyStart-1] == ']') {
	//	return "", false
	//}

	valStart := keyStart + len(confHolder)
	valEnd := strings.Index(conf[valStart:], "]")
	if valEnd <= 0 {
		return "", false, FigError{
			Cause:  "Invalid configuration in: " + conf + " for configuration: " + key,
			Error_: ErrorIncorrectTagConfiguration,
		}
	}
	return string(conf[valStart : valStart+valEnd]), true, nil
}

type Fig struct {
	injectOnlyIfFigTagProvided bool
	registered                 map[reflect.Type]interface{}
	assembled                  map[reflect.Type]bool

	registeredValues map[string]interface{}

	assembleRegisteredOnce *sync.Once
}

func New(injectOnlyIfFigTagProvided bool) *Fig {
	return &Fig{
		injectOnlyIfFigTagProvided: injectOnlyIfFigTagProvided,
		registered:                 make(map[reflect.Type]interface{}),
		assembled:                  make(map[reflect.Type]bool),
		registeredValues:           make(map[string]interface{}),
		assembleRegisteredOnce:     new(sync.Once),
	}
}

var (
	ErrorCannotBeRegistered         = errors.New("provided value can't be registered")
	ErrorCannotBeHolder             = errors.New("provided value can't be holder")
	ErrorCannotDecideImplementation = errors.New("not able to get value to inject")
	ErrorRegisteredValueOverridden  = errors.New("already registered value was overridden")
	ErrorIncorrectTagConfiguration  = errors.New("invalid `fig` tag configuration")
)

type FigError struct {
	Error_ error
	Cause  string
}

func (fe FigError) Error() string {
	return fmt.Sprintf("Message: %s. Cause: %s", fe.Cause, fe.Error_.Error())
}

func (fig *Fig) Register(impls ...interface{}) error {
	for _, impl := range impls {
		implType := reflect.TypeOf(impl)
		if implType == nil {
			return FigError{Cause: "nil cannot be registered as injectable value", Error_: ErrorCannotBeRegistered}
		}

		if implType.Kind() == reflect.Struct ||
			implType.Kind() == reflect.Ptr && implType.Elem().Kind() == reflect.Struct {
			fig.registered[implType] = impl
		} else {
			return FigError{Cause: "only structs and references to structs can be registered", Error_: ErrorCannotBeRegistered}
		}
	}
	return nil
}

func (fig *Fig) RegisterValue(key string, value interface{}) error {
	if value == nil {
		return FigError{
			Error_: ErrorCannotBeRegistered,
		}
	}
	if _, found := fig.registeredValues[key]; found {
		fig.registeredValues[key] = value
		return FigError{
			Error_: ErrorRegisteredValueOverridden,
		}
	}
	fig.registeredValues[key] = value
	return nil
}

func (fig *Fig) RegisterValues(keyValues map[string]interface{}) error {
	for key, value := range keyValues {
		if err := fig.RegisterValue(key, value); err != nil {
			return err
		}
	}
	return nil
}

func (fig *Fig) Initialize(holder interface{}) error {
	holderType := reflect.TypeOf(holder)
	if holderType == nil {
		return FigError{Cause: "nil cannot be holder", Error_: ErrorCannotBeHolder}
	}
	if holderType.Kind() != reflect.Ptr ||
		holderType.Elem().Kind() != reflect.Struct {
		return FigError{Cause: "only references to structs can be holders", Error_: ErrorCannotBeHolder}
	}
	if err := fig.AssembleRegistered(); err != nil {
		return err
	}
	return fig.assemble(holder, false)
}

func isCircleInclusion(this reflect.Type, that reflect.Type) bool {
	for this.Kind() == reflect.Ptr {
		this = this.Elem()
	}
	if this.Kind() == reflect.Struct {
		numFields := this.NumField()
		for i := 0; i < numFields; i++ {
			thisFieldType := this.Field(i).Type
			if that.AssignableTo(thisFieldType) {
				return true
			}
			if isCircleInclusion(thisFieldType, that) {
				return true
			}
		}
	}
	return false
}

func (fig *Fig) AssembleRegistered() error {
	var err error
	fig.assembleRegisteredOnce.Do(func() {
		for regType, regObject := range fig.registered {
			if err = fig.assemble(regObject, true); err != nil {
				break
			}
			fig.assembled[regType] = true
		}
	})
	return err
}

func getFigTagConfig(tag reflect.StructTag, key string) (string, bool, error) {
	if figTag, ok := tag.Lookup(FIG_TAG); ok {
		return getConfigValueForKey(figTag, key)
	} else {
		return "", false, nil
	}
}

func setByImplConf(canBeSet []interface{}, elementField reflect.Value, implFigConf string) error {
	for _, canBe := range canBeSet {
		implName := getFullName(canBe)
		if implName == implFigConf {
			elementField.Addr().Elem().Set(reflect.ValueOf(canBe))
			return nil
		}
	}
	return FigError{
		Cause:  fmt.Sprintf("implementation defined in tag was not found: %s", implFigConf),
		Error_: ErrorCannotDecideImplementation,
	}
}

func setByQualConf(canBeSet []interface{}, elementField reflect.Value, qualFigConf string) error {
	for _, canBe := range canBeSet {
		if checkQualifier(canBe, qualFigConf) {
			elementField.Addr().Elem().Set(reflect.ValueOf(canBe))
			return nil
		}
	}
	return FigError{
		Cause:  fmt.Sprintf("condition defined in tag was not found: %s", qualFigConf),
		Error_: ErrorCannotDecideImplementation,
	}
}

func (fig *Fig) setFoundImpl(canBeSet []interface{}, elementField reflect.Value, tag reflect.StructTag) error {
	switch {
	case len(canBeSet) > 1:
		if implFigConf, found, err := getFigTagConfig(tag, IMPL_TAG_KEY); err != nil {
			return err
		} else if found {
			return setByImplConf(canBeSet, elementField, implFigConf)
		} else if qualFigConf, found, err := getFigTagConfig(tag, QUAL_TAG_KEY); err != nil {
			return err
		} else if found {
			return setByQualConf(canBeSet, elementField, qualFigConf)
		} else {
			mes := "can't chose implementation for " + elementField.String() + ":\n"
			for _, canBe := range canBeSet {
				mes += fmt.Sprintf("\t%T\n", canBe)
			}
			return FigError{Cause: mes, Error_: ErrorCannotDecideImplementation}
		}
	case len(canBeSet) < 1:
		switch elementField.Kind() {
		case reflect.Struct:
			if err := fig.Initialize(elementField.Addr().Interface()); err != nil {
				return err
			}
		case reflect.Ptr:
			if elementField.IsNil() {
				elementField.Set(reflect.New(elementField.Type().Elem()))
			}
			if err := fig.Initialize(elementField.Interface()); err != nil {
				return err
			}
		default:
			return FigError{Cause: "no implementation found for " + elementField.String(), Error_: ErrorCannotDecideImplementation}
		}
	default:
		elementField.Addr().Elem().Set(reflect.ValueOf(canBeSet[0]))
	}
	return nil
}

func checkQualifier(canBe interface{}, qualFigConf string) bool {
	canBeType := reflect.TypeOf(canBe)
	if canBeType.Implements(reflect.TypeOf((*Qualifier)(nil)).Elem()) {
		if canBe.(Qualifier).Qualify() == qualFigConf {
			return true
		}
	}
	return false
}

func getFullName(canBe interface{}) string {
	var implName string
	var pckPath string
	typeOfCanBe := reflect.TypeOf(canBe)
	if typeOfCanBe.Kind() == reflect.Ptr && typeOfCanBe.Elem().Kind() == reflect.Struct {
		pckPath = typeOfCanBe.Elem().PkgPath()
		implName = typeOfCanBe.Elem().Name()
	} else if typeOfCanBe.Kind() == reflect.Struct {
		pckPath = typeOfCanBe.PkgPath()
		implName = typeOfCanBe.Name()
	} else {
		return ""
	}
	if pckPath != "" {
		implName = pckPath + "/" + implName
	}
	return implName
}

type InjectStep interface {
	Do() error
	Break() bool
}

type InjectStepValueSetup struct {
	fig                *Fig
	holderElementField reflect.Value
	tag                reflect.StructTag
	recursive          bool
}

func NewValueSetup(fig *Fig, tag reflect.StructTag, holderElementField reflect.Value, recursive bool) *InjectStepValueSetup {
	return &InjectStepValueSetup{fig: fig, holderElementField: holderElementField, recursive: recursive, tag: tag}
}

func (valueSetup *InjectStepValueSetup) Do() error {
	switch valueSetup.holderElementField.Kind() {
	case reflect.Interface:
		var canBeSet []interface{}
		for registeredType, injectableObj := range valueSetup.fig.registered {
			if registeredType.Implements(valueSetup.holderElementField.Type()) &&
				!isCircleInclusion(registeredType, valueSetup.holderElementField.Type()) {
				if valueSetup.recursive && !valueSetup.fig.assembled[registeredType] {
					if err := valueSetup.fig.assemble(injectableObj, valueSetup.recursive); err != nil {
						return err
					}
				}
				canBeSet = append(canBeSet, injectableObj)
			}
		}
		if err := valueSetup.fig.setFoundImpl(canBeSet, valueSetup.holderElementField, valueSetup.tag); err != nil {
			return err
		}

	case reflect.Struct:
		var canBeSet []interface{}
		for registeredType, injectableObj := range valueSetup.fig.registered {
			if registeredType.AssignableTo(valueSetup.holderElementField.Type()) {
				if valueSetup.recursive && !valueSetup.fig.assembled[registeredType] {
					if err := valueSetup.fig.assemble(injectableObj, valueSetup.recursive); err != nil {
						return err
					}
				}
				canBeSet = append(canBeSet, injectableObj)
			}
		}
		if err := valueSetup.fig.setFoundImpl(canBeSet, valueSetup.holderElementField, valueSetup.tag); err != nil {
			return err
		}

	case reflect.Ptr:
		for valueSetup.holderElementField.Kind() == reflect.Ptr {
			valueSetup.holderElementField.Set(reflect.New(valueSetup.holderElementField.Type().Elem()))
			valueSetup.holderElementField = valueSetup.holderElementField.Elem()
		}
		if err := valueSetup.Do(); err != nil {
			return err
		}

	case
		reflect.String:
		if envKey, found, err := getFigTagConfig(valueSetup.tag, ENV_TAG_KEY); err != nil {
			return err
		} else if found {
			envVal := os.Getenv(envKey)
			valueSetup.holderElementField.Addr().Elem().Set(reflect.ValueOf(envVal))
		}

	//case
	//	reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
	//	reflect.Float32, reflect.Float64,
	//	reflect.Bool,
	//	reflect.Slice,
	//	reflect.Map,
	//	reflect.Complex64, reflect.Complex128,
	//	reflect.Array,
	//	reflect.Chan,
	//	reflect.Func:
	//	return FigError{Cause: "Unsupported holder field type: " + valueSetup.holderElementField.String(), Error_: ErrorCannotBeHolder}
	default:
		return FigError{Cause: "Unsupported holder field type: " + valueSetup.holderElementField.String(), Error_: ErrorCannotBeHolder}
	}
	return nil
}

func (valueSetup *InjectStepValueSetup) Break() bool {
	return true
}

type InjectStepRegisteredValueSetup struct {
	fig                *Fig
	tag                reflect.StructTag
	holderElementField reflect.Value
	skip               bool
}

func NewRegisteredValueSetup(fig *Fig, tag reflect.StructTag, holderElementField reflect.Value) *InjectStepRegisteredValueSetup {
	return &InjectStepRegisteredValueSetup{fig: fig, tag: tag, holderElementField: holderElementField}
}

func (registeredValue *InjectStepRegisteredValueSetup) Do() error {
	if regKey, found, err := getFigTagConfig(registeredValue.tag, REG_TAG_KEY); err != nil {
		return err
	} else if found {
		if regValue, found := registeredValue.fig.registeredValues[regKey]; found {
			registeredValue.holderElementField.Addr().Elem().Set(reflect.ValueOf(regValue))
			registeredValue.skip = true
		} else {
			return FigError{
				Cause:  fmt.Sprintf("Implementation was not found for: %s", registeredValue.holderElementField),
				Error_: ErrorCannotDecideImplementation,
			}
		}
	}
	return nil
}

func (registeredValue *InjectStepRegisteredValueSetup) Break() bool {
	return registeredValue.skip
}

type InjectStepFigTagRequiredCheck struct {
	fig  *Fig
	tag  reflect.StructTag
	skip bool
}

func NewFigTagRequiredCheck(fig *Fig, tag reflect.StructTag) *InjectStepFigTagRequiredCheck {
	return &InjectStepFigTagRequiredCheck{fig: fig, tag: tag}
}

func (figTagRequired *InjectStepFigTagRequiredCheck) Do() error {
	if figTagRequired.fig.injectOnlyIfFigTagProvided {
		if _, found := figTagRequired.tag.Lookup(FIG_TAG); !found {
			figTagRequired.skip = true
		}
	}
	return nil
}

func (figTagRequired *InjectStepFigTagRequiredCheck) Break() bool {
	return figTagRequired.skip
}

type InjectStepSkipCheck struct {
	tag reflect.StructTag

	skip bool
}

func NewSkipCheck(tag reflect.StructTag) *InjectStepSkipCheck {
	return &InjectStepSkipCheck{tag: tag}
}

func (skipVerify *InjectStepSkipCheck) Do() error {
	if skipConfValue, found, err := getFigTagConfig(skipVerify.tag, SKIP_TAG_KEY); err != nil {
		return err
	} else if found {
		if skipConfValue == "true" {
			skipVerify.skip = true
		} else if skipConfValue != "false" {
			return FigError{
				Cause:  "Incorrectly defined configuration `skip` of `fig` tag. Supported values: true | false. Got: " + skipConfValue,
				Error_: ErrorIncorrectTagConfiguration,
			}
		}
	}
	return nil
}

func (skipVerify *InjectStepSkipCheck) Break() bool {
	return skipVerify.skip
}

type StepMachine struct {
	steps []InjectStep
}

func NewStepMachine() *StepMachine {
	return &StepMachine{}
}

func (sm *StepMachine) Add(steps ...InjectStep) {
	sm.steps = append(sm.steps, steps...)
}

func (sm *StepMachine) Do() error {
	for _, step := range sm.steps {
		if err := step.Do(); err != nil {
			return err
		}
		if step.Break() {
			break
		}
	}
	return nil
}

func (fig *Fig) assemble(holder interface{}, recursive bool) error {
	holderElement := reflect.ValueOf(holder)
	if holderElement.Kind() == reflect.Ptr {
		holderElement = holderElement.Elem()
	}
	holderElementType := holderElement.Type()
	numFields := holderElement.NumField()

	for fieldIndex := 0; fieldIndex < numFields; fieldIndex++ {
		holderElementField := holderElement.Field(fieldIndex)
		tag := holderElementType.Field(fieldIndex).Tag

		stepMachine := NewStepMachine()
		stepMachine.Add(
			NewFigTagRequiredCheck(fig, tag),
			NewSkipCheck(tag),
			NewRegisteredValueSetup(fig, tag, holderElementField),
			NewValueSetup(fig, tag, holderElementField, recursive),
		)
		if err := stepMachine.Do(); err != nil {
			return err
		}
	}
	return nil
}
