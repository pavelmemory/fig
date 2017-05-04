package di

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func getConfigValueForKey(conf string, key string) (string, bool) {
	key += "["
	keyStart := strings.Index(conf, key)
	if keyStart < 0 {
		return "", false
	}
	if keyStart > 0 && !(conf[keyStart-1] == ' ' || conf[keyStart-1] == ']') {
		return "", false
	}

	valStart := keyStart + len(key)
	searchFrom := valStart
	var valEnd int
	for {
		valEnd = strings.Index(conf[searchFrom:], "]")
		if valEnd <= 0 {
			return "", false
		}
		if conf[valEnd-1] == '\\' {
			searchFrom += valEnd + 1
		} else {
			return string(conf[valStart : valEnd+valStart]), true
		}
	}
}

type Fig struct {
	registered map[reflect.Type]interface{}
	assembled  map[reflect.Type]bool

	assembleRegisteredOnce *sync.Once
}

func New() *Fig {
	return &Fig{
		registered:             make(map[reflect.Type]interface{}),
		assembled:              make(map[reflect.Type]bool),
		assembleRegisteredOnce: new(sync.Once),
	}
}

var (
	ErrorCannotBeRegister           = errors.New("provided value can't be registered")
	ErrorCannotBeHolder             = errors.New("provided value can't be holder")
	ErrorCannotDecideImplementation = errors.New("holder must have explicit implementation defined")
)

type FigError struct {
	err   error
	Cause string
}

func (fe FigError) Error() string {
	return fmt.Sprintf("Message: %s. Cause: %s", fe.Cause, fe.err.Error())
}

func (fig *Fig) Register(impls ...interface{}) error {
	for _, impl := range impls {
		implType := reflect.TypeOf(impl)
		if implType == nil {
			return FigError{Cause: "you cannot register nil", err: ErrorCannotBeRegister}
		}

		if implType.Kind() == reflect.Struct ||
			implType.Kind() == reflect.Ptr && implType.Elem().Kind() == reflect.Struct {
			fig.registered[implType] = impl
		} else {
			return FigError{Cause: "only structs and references to structs can be registered", err: ErrorCannotBeRegister}
		}
	}
	return nil
}

func (fig *Fig) Initialize(holder interface{}) error {
	holderType := reflect.TypeOf(holder)
	if holderType == nil {
		return FigError{Cause: "nil cannot be holder", err: ErrorCannotBeHolder}
	}
	if holderType.Kind() != reflect.Ptr ||
		holderType.Elem().Kind() != reflect.Struct {
		return FigError{Cause: "only references to structs can be holders", err: ErrorCannotBeHolder}
	}
	if err := fig.AssembleRegistered(); err != nil {
		return err
	}
	return fig.assemble(holder, false)
}

func isCircleInclusion(this reflect.Type, that reflect.Type) bool {
	if this.Kind() == reflect.Ptr {
		this = this.Elem()
	}
	if this.Kind() == reflect.Struct {
		numFields := this.NumField()
		for i := 0; i < numFields; i++ {
			thisFieldType := this.Field(i).Type
			fmt.Println(thisFieldType.String(), that.String())
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
				return
			}
			fig.assembled[regType] = true
		}
	})
	return err
}

func getFigTagConfig(tag reflect.StructTag, key string) (string, bool) {
	if figTag, ok := tag.Lookup("fig"); ok {
		return getConfigValueForKey(figTag, key)
	} else {
		return "", false
	}
}

func (fig *Fig) setFoundImpl(canBeSet []interface{}, elementField reflect.Value, tag reflect.StructTag) error {
	if len(canBeSet) > 1 {
		implFigConf, found := getFigTagConfig(tag, "impl")
		if !found {
			mes := "Can't chose implementation for " + elementField.String() + ". Found:\n"
			for _, canBe := range canBeSet {
				mes += fmt.Sprintf("\t%T\n", canBe)
			}
			return errors.New(mes)
		}
		for _, canBe := range canBeSet {
			implName := getFullName(canBe)
			if implName == implFigConf {
				elementField.Addr().Elem().Set(reflect.ValueOf(canBe))
				return nil
			}
		}
		return errors.New(fmt.Sprintf("Implementation specification defined in tag not found: %s", implFigConf))
	} else if len(canBeSet) < 1 {
		return errors.New("No implementation found for " + elementField.String())
	} else {
		elementField.Addr().Elem().Set(reflect.ValueOf(canBeSet[0]))
		return nil
	}
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

func (fig *Fig) assemble(holder interface{}, recursive bool) error {
	holderElement := reflect.ValueOf(holder)
	if holderElement.Kind() == reflect.Ptr {
		holderElement = holderElement.Elem()
	}
	holderElementType := holderElement.Type()
	numFields := holderElement.NumField()
	allFields := make(map[reflect.Value]struct{})

	for fieldIndex := 0; fieldIndex < numFields; fieldIndex++ {
		holderElementField := holderElement.Field(fieldIndex)
		holderElementFieldType := holderElementField.Type()
		tag := holderElementType.Field(fieldIndex).Tag
		if skipConfValue, found := getFigTagConfig(tag, "skip"); found {
			if needSkip, err := strconv.ParseBool(skipConfValue); err != nil {
				return err
			} else if needSkip {
				continue
			}
		}
		allFields[holderElementField] = struct{}{}

		switch holderElementField.Kind() {
		case reflect.Interface:
			var canBeSet []interface{}
			for registeredType, injectableObj := range fig.registered {
				if registeredType.Implements(holderElementFieldType) && !isCircleInclusion(registeredType, holderElementFieldType) {
					if recursive && !fig.assembled[registeredType] {
						if err := fig.assemble(injectableObj, recursive); err != nil {
							return err
						}
					}
					canBeSet = append(canBeSet, injectableObj)
				}
			}
			if err := fig.setFoundImpl(canBeSet, holderElementField, tag); err != nil {
				return err
			}

		case reflect.Ptr:
			// TODO: we want be able to set basic values with reference type(optional parameters)
			if holderElementFieldType.Elem().Kind() == reflect.Struct {
				var canBeSet []interface{}
				for registeredType, injectableObj := range fig.registered {
					if registeredType.AssignableTo(holderElementFieldType) {
						if recursive && !fig.assembled[registeredType] {
							if err := fig.assemble(injectableObj, recursive); err != nil {
								return err
							}
						}
						canBeSet = append(canBeSet, injectableObj)
					}
				}
				if err := fig.setFoundImpl(canBeSet, holderElementField, tag); err != nil {
					return err
				}
			} else {
				return FigError{Cause: "Can't inject to non-struct reference fields", err: ErrorCannotBeHolder}
			}

		case reflect.Struct:
			var canBeSet []interface{}
			for registeredType, injectableObj := range fig.registered {
				if registeredType.AssignableTo(holderElementFieldType) {
					if recursive && !fig.assembled[registeredType] {
						if err := fig.assemble(injectableObj, recursive); err != nil {
							return err
						}
					}
					canBeSet = append(canBeSet, injectableObj)
				}
			}
			if err := fig.setFoundImpl(canBeSet, holderElementField, tag); err != nil {
				return err
			}

		case
			reflect.String:
			if envKey, found := getFigTagConfig(tag, "env"); found {
				envVal := os.Getenv(envKey)
				holderElementField.Addr().Elem().Set(reflect.ValueOf(envVal))
			}

		case
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Float32, reflect.Float64,
			reflect.Bool,
			reflect.Slice,
			reflect.Map,
			reflect.Complex64, reflect.Complex128,
			reflect.Array,
			reflect.Chan,
			reflect.Func:
			return FigError{Cause: "Unsupported holder field type: " + holderElementField.String(), err: ErrorCannotBeHolder}
		default:
			return FigError{Cause: "Unsupported holder field type: " + holderElementField.String(), err: ErrorCannotBeHolder}
		}
		delete(allFields, holderElementField)
	}
	if len(allFields) > 0 {
		var notInjected []string
		for field := range allFields {
			notInjected = append(notInjected, field.String())
		}
		return FigError{
			Cause: fmt.Sprintf("Implementations were not found for: %s", strings.Join(notInjected, ", ")),
			err:   ErrorCannotDecideImplementation,
		}
	}
	return nil
}
