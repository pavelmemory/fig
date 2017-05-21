package simple

import (
	"fmt"
	"os"

	"github.com/pavelmemory/fig"
)

type UserRepo interface {
	Find() []string
}

type MemUserRepo struct {
	message string
}

func (mur *MemUserRepo) Find() []string {
	return []string{"mem", mur.message}
}

type FileUserRepo struct {
	message string
}

func (fur *FileUserRepo) Find() []string {
	return []string{"file", fur.message}
}

type OrderRepo struct {
}

func (or *OrderRepo) Create() string {
	return "order created"
}

type Module struct {
	Name      string `fig:"env[ENV_NAME]"`
	OrderRepo *OrderRepo
	UserRepo
}

func init() {
	os.Setenv("ENV_NAME", "DEV")
}

func ExampleOfSimpleInjectionOnDependenciesAndEnvVar() {
	injector := fig.New(false)
	err := injector.Register(
		&MemUserRepo{message: "mes_1"},
	)
	if err != nil {
		panic(fmt.Sprintf("%#v", err))
	}

	module := new(Module)

	err = injector.Initialize(module)
	if err != nil {
		panic(fmt.Sprintf("%#v", err))
	}

	fmt.Println(module.Name)
	fmt.Println(module.Find())
	fmt.Println(module.OrderRepo.Create())
	// Output:
	//DEV
	//[mem mes_1]
	//order created
}
