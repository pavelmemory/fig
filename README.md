## Fig
**Fig** is a dependency injection tool written in Go for Go.
It was developed to help solve problem with initialization of structures
with dependencies with dependencies with depend...

It is simple to use, lightweight and written in pure Go.

Common scenario when to use it is next. Lets assume you have a struct with fields of
interface types and you want them to be initialized with right implementation
depending if it is a prod, dev or test run. All you need to do is
create Fig object and register to it all dependencies you would like to
use. It is not necessary to initialize dependencies of dependencies because it will be done
automatically. All registered objects will be treated as potential candidates
for injection. After that just call `Initialize` method with struct you want to
be initialized (it also can be reference to anonymous struct).
Please be sure that you have registered reference to object
if it's methods expect reference receiver.

***Example***

Lets assume we have next `main.go` file with some structs and we want to avoid 
manual assignments of all dependencies, so you can simply 
inject them.
```go
package main

import (
    "fmt"
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

func main() {
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
}
```
The result of execution `ENV_NAME=DEV go run main.go` is:
```text
DEV
[mem mes_1]
order created
```

There are two modes how `Fig` works. In case you create `Fig` object
with `false` constructor value will mean you want all fields
of initialization structs to be injected. If you specify `true`
 then only fields with `fig` tag will be injected (it can be just empty
 tag like `fig:""`).

**Example**
'fig' injector with 'false'
https://github.com/pavelmemory/fig/blob/master/examples/sample/sample_2_test.go
**Example**
'fig' injector with 'true'
https://github.com/pavelmemory/fig/blob/master/examples/sample/sample_3_test.go


****
**Multiple implementations of interface**

In case you have registered multiple implementations of one interface 
you need to define a policy that will help to find which implementation to inject.
There are couple of configuration tags designed to help with it.
You need to configure fields with tag `fig` and policies from list below
This tag can have next configurations:
- `skip` - expected value [`true`|`false`].
This field will be skipped at time of injection
if provided value is `true` and the field 
has value it had before invocation of `Initialize`
method (can be used with other configurations)
- `impl` - expected value is a full name of type with
package name `github.com/proj/implement/SomeStructName`
It will be used in case you have registered multiple
implementations of one interface. So object with provided
name will be injected
- `qual` - expected value is any string. It will be used
in case you have registered multiple implementations
of one interface. Implementation will be chosen based on
value provided in this config and value that will be
returned by `Qualify() string` method of one of
registered implementations. If no one of candidates implements
`Qualifier` interface or no equal to specified string returned
it will lead to an error

***Example***
https://github.com/pavelmemory/fig/blob/master/examples/sample/sample_2_test.go

****
**Any value by key**

You can use `reg` configuration of `fig` tag to inject anything you want.
You should register all values you want to use with
`RegisterValue(key string, value interface{}) error` or 
`RegisterValues(keyValues map[string]interface{}) error` methods.
For configuration `fig:"reg[key1]"` value associated with key `key1` 
will be injected to the field.
- `reg` - expected value is any string. Must be
registered with `RegisterValue` or `RegisterValues` methods

***Example***
https://github.com/pavelmemory/fig/blob/master/examples/sample/sample_5_test.go

***
**Environment variables injection**

Scenario when you need some environment variable value is quite
popular, so there is solution for that. Tag configuration `env` solves
the problem of manual initialization of such values.
- `env` - expected value is any string that can represent key of
environment variable. Value of this environment variable will be assigned to field.
There is no way to provide default value, so in such case it is better to use `reg`
configuration

***Example***
https://github.com/pavelmemory/fig/blob/master/examples/sample/sample_1_test.go

***
**Initialization of maps, slices and channels**

Maps, slices and channels can be initialized as well.
<br/>For maps just new map struct will be created and injected.
<br/>For channels it is possible to define size of buffer. Default value is 1.
<br/>And for slices you can define length and capacity. Default length and capacity are 1.
- `size` - expected value is integer. Used to specify size of channel or 
length of slice.
- `cap` - expected value is integer. Used to specify capacity of slice. Can't be less than `size` for slices.

***Example***
https://github.com/pavelmemory/fig/blob/master/examples/sample/sample_5_test.go
