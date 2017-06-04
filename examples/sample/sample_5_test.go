package sample

import (
	"fmt"

	"github.com/pavelmemory/fig"
)

func ExampleInjectionByRegisteredKeyValuePair() {
	injector := fig.New(false)

	usefulStruct := struct {
		Port  int      `fig:"reg[port]"`
		Angle *float64 `fig:"reg[0_angle]"`
	}{}

	FatalIfError(func() error {
		return injector.RegisterValue("port", 9999)
	})

	FatalIfError(func() error {
		angle := 0.64
		return injector.RegisterValues(map[string]interface{}{"0_angle": &angle})
	})

	FatalIfError(func() error {
		return injector.Initialize(&usefulStruct)
	})

	fmt.Println("Port is", usefulStruct.Port)
	fmt.Println("Angle is", *usefulStruct.Angle)
	// Output:
	// Port is 9999
	// Angle is 0.64
}
