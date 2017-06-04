package sample

import (
	"fmt"

	"github.com/pavelmemory/fig"
)

func ExampleInjectionOfMapsSlicesAndChannels() {
	injector := fig.New(false)

	usefulStruct := struct {
		Limiter      chan struct{}
		InboundQueue chan string `fig:"size[256]"`
		List         []int       `fig:"cap[16]"`
	}{}

	FatalIfError(func() error {
		return injector.Initialize(&usefulStruct)
	})

	fmt.Println("Limiter capacity is", cap(usefulStruct.Limiter))
	fmt.Println("InboundQueue channel's buffer size is", cap(usefulStruct.InboundQueue))
	fmt.Println("List's length is", len(usefulStruct.List), "and capacity is", cap(usefulStruct.List))
	// Output:
	// Limiter capacity is 1
	// InboundQueue channel's buffer size is 256
	// List's length is 0 and capacity is 16
}
