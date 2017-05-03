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

func (fig *Fig) Register(impls ...interface{}) error {
	for _, impl := range impls {
		tpe := reflect.TypeOf(impl)
		if tpe == nil {
			return errors.New("nil reference can't be used")
		}

		if tpe.Kind() == reflect.Struct ||
			tpe.Kind() == reflect.Ptr && tpe.Elem().Kind() == reflect.Struct {
			fig.registered[tpe] = impl
		} else {
			return errors.New("you can register only structs and references to structs")
		}
	}
	return nil
}

func AsString(strct interface{}) string {
	// TODO: fmt.Sprintf("%T", svla)
	//if refVal, ok := strct.(reflect.Value); ok {
	//	return refVal.String()
	//}
	//
	//if refType, ok := strct.(reflect.Type); ok {
	//	return refType.String()
	//}
	return fmt.Sprintf("%T", strct)
}

func (fig *Fig) Initialize(strct interface{}) error {
	if reflect.TypeOf(strct).Kind() != reflect.Ptr {
		return errors.New("Accept only references")
	}
	if err := fig.AssembleRegistered(); err != nil {
		return err
	}
	return fig.assemble(strct, 0, false)
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
		for regType, regStruct := range fig.registered {
			if err = fig.assemble(regStruct, 0, true); err != nil {
				return
			}
			fig.assembled[regType] = true
		}
	})
	return err
}

func getFigTagConfig(holder interface{}, fieldIndex int, key string) (string, bool) {
	tag := getTag(reflect.TypeOf(holder), fieldIndex)
	if figTag, ok := tag.Lookup("fig"); ok {
		return getConfigValueForKey(figTag, key)
	} else {
		return "", false
	}
}

func (fig *Fig) setFoundImpl(canBeSet []interface{}, holder interface{}, fieldIndex int) (string, error) {
	holderValue := reflect.ValueOf(holder)
	if holderValue.Kind() == reflect.Ptr {
		if holderValue.Elem().Kind() == reflect.Struct {
			holderValue = holderValue.Elem()
		} else {
			return "", errors.New("Not able to inject to non struct objects")
		}
	} else if holderValue.Kind() != reflect.Struct {
		return "", errors.New("Not able to inject to non struct objects")
	}

	elementField := holderValue.Field(fieldIndex)
	elementFieldType := elementField.Type()

	if skipConfValue, found := getFigTagConfig(holder, fieldIndex, "skip"); found {
		if needSkip, err := strconv.ParseBool(skipConfValue); err != nil {
			return "", err
		} else if needSkip {
			return elementFieldType.Name(), nil
		}
	}

	if len(canBeSet) > 1 {
		fmt.Println("Decision between multiple implementations:", canBeSet)
		implFigConf, found := getFigTagConfig(holder, fieldIndex, "impl")
		if !found {
			mes := "Can't chose implementation for " + AsString(elementField) + ". Found:\n"
			for _, canBe := range canBeSet {
				mes += fmt.Sprintf("\t%T\n", canBe)
			}
			return "", errors.New(mes)
		}
		for _, canBe := range canBeSet {
			implName := getFullName(canBe)
			if implName == implFigConf {
				elementField.Addr().Elem().Set(reflect.ValueOf(canBe))
				return elementFieldType.Name(), nil
			}
		}
		return "", errors.New(fmt.Sprintf("Implementation specification defined in tag not found: %s", implFigConf))
	} else if len(canBeSet) < 1 {
		return "", errors.New("No implementation found for " + AsString(elementField))
	} else {
		elementField.Addr().Elem().Set(reflect.ValueOf(canBeSet[0]))
		return elementFieldType.Name(), nil
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

func getTag(holderType reflect.Type, fieldIndex int) reflect.StructTag {
	if holderType.Kind() == reflect.Ptr && holderType.Elem().Kind() == reflect.Struct {
		return holderType.Elem().Field(fieldIndex).Tag
	} else {
		return holderType.Field(fieldIndex).Tag
	}
}

func (fig *Fig) assemble(holder interface{}, deep int, recursive bool) error {
	fmt.Println("Start assembling of", AsString(holder))
	if holder == nil {
		return errors.New("nil reference can't be used")
	}
	holderValue := reflect.ValueOf(holder)
	if err := validateHolder(holderValue); err != nil {
		return err
	}

	var holderElement reflect.Value
	if holderValue.Kind() == reflect.Ptr {
		holderElement = holderValue.Elem()
	} else {
		holderElement = holderValue
	}
	numField := holderElement.NumField()
	allFields := make(map[string]struct{})

	for fieldIndex := 0; fieldIndex < numField; fieldIndex++ {
		elementField := holderElement.Field(fieldIndex)
		fmt.Println("Start assembling of field", AsString(elementField))
		elementFieldType := elementField.Type()
		allFields[elementFieldType.Name()] = struct{}{}

		switch elementField.Kind() {
		case reflect.Interface:
			fmt.Println("Assembling of interface field")
			var canBeSet []interface{}
			for registeredType, injectableObj := range fig.registered {
				if registeredType.Implements(elementFieldType) && !isCircleInclusion(registeredType, elementFieldType) {
					if recursive && !fig.assembled[registeredType] {
						if err := fig.assemble(injectableObj, deep+1, recursive); err != nil {
							return err
						}
					}
					canBeSet = append(canBeSet, injectableObj)
					fmt.Println("Potential candidate: ", AsString(injectableObj))
				}
			}

			if elementFieldTypeName, err := fig.setFoundImpl(canBeSet, holder, fieldIndex); err != nil {
				return err
			} else {
				delete(allFields, elementFieldTypeName)
			}
		case reflect.Ptr:
			// TODO: we want be able to set basic values with reference type(optional parameters)
			if elementFieldType.Elem().Kind() == reflect.Struct {
				fmt.Println("Assembling of struct(reference) field")
				var canBeSet []interface{}
				for registeredType, injectableObj := range fig.registered {
					if registeredType.AssignableTo(elementFieldType) {
						fmt.Println(registeredType.String(), "is AssignableTo", elementFieldType.String(), deep)
						if recursive && !fig.assembled[registeredType] {
							if err := fig.assemble(injectableObj, deep+1, recursive); err != nil {
								return err
							}
						}
						canBeSet = append(canBeSet, injectableObj)
						fmt.Println("Potential candidate: ", AsString(injectableObj))
					}
				}

				if elementFieldTypeName, err := fig.setFoundImpl(canBeSet, holder, fieldIndex); err != nil {
					return err
				} else {
					delete(allFields, elementFieldTypeName)
				}
			} else {
				return errors.New("Can't inject non-struct reference fields")
			}
		case reflect.Struct:
			fmt.Println("Assembling of struct(value) field")
			var canBeSet []interface{}
			for registeredType, injectableObj := range fig.registered {
				if registeredType.AssignableTo(elementFieldType) {
					fmt.Println(registeredType.String(), "is AssignableTo", elementFieldType.String(), deep)
					if recursive && !fig.assembled[registeredType] {
						if err := fig.assemble(injectableObj, deep+1, recursive); err != nil {
							return err
						}
					}
					canBeSet = append(canBeSet, injectableObj)
					fmt.Println("Potential candidate: ", AsString(injectableObj))
				}
			}

			if elementFieldTypeName, err := fig.setFoundImpl(canBeSet, holder, fieldIndex); err != nil {
				return err
			} else {
				delete(allFields, elementFieldTypeName)
			}
		case
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Float32, reflect.Float64,
			reflect.Bool,
			reflect.Slice,
			reflect.Map,
			reflect.String:

			if envKey, found := getFigTagConfig(holder, fieldIndex, "env"); found {
				envVal := os.Getenv(envKey)
				elementField.Addr().Elem().Set(reflect.ValueOf(envVal))
			}
			delete(allFields, elementFieldType.Name())
			fmt.Println("Basic type assembling:", elementField.Type().String(), elementFieldType.Name(), deep)

		case
			reflect.Complex64, reflect.Complex128,
			reflect.Array,
			reflect.Chan,
			reflect.Func:
			delete(allFields, elementFieldType.Name())
			fmt.Println("Basic type assembling: ", elementFieldType.Name(), deep)
		default:
			fmt.Println("Type assembling: ", elementFieldType.Name(), deep)
		}
	}
	if len(allFields) > 0 {
		var notInjected []string
		for fieldName := range allFields {
			notInjected = append(notInjected, fieldName)
		}
		return errors.New(fmt.Sprintf("Implementations were not found for: %s %d", strings.Join(notInjected, ", "), deep))
	}
	return nil
}

func validateHolder(value reflect.Value) error {
	vkind := value.Kind()
	if vkind == reflect.Struct ||
		(vkind == reflect.Ptr && value.Elem().Kind() == reflect.Struct) {
		return nil
	}
	return errors.New("you can use only structs and references to structs as points of injection")
}
