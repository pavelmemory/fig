All registered objects will be treated as potential candidates
for injection.
Please be sure that you have registered reference to object
if it methods expect reference receiver.
Default scenario to use it: you have struct with fields of
interface types, you need to register entities that implement
those interfaces, so thay can be injected with Initialize method.

There is two modes how `Fig` works. In case you create `Fig` object
with `false` constructor value. It means you want that all fields
of initialization structs to be injected. If you specify `true`
 than only fields with `fig` tag will be injected(it can be just empty
 tag like `fig:""`).

***
Example

TODO: example for each of explained configuration


****
Multiple implementations of interface

In case you have multiple implementations registered to be injected and need define some policy what you whant in each injectable field.
There is a couple of configuration tags that can help with it.
To use those configurations you need to define tag 'fig' for field.
This tag can have next configurations:
- skip - expected value [`true`|`false`].
This field will be skipped in time of injection
if provided value will be `true` and the field will
has value it has before invocation of Initialize
method(can be used with other configurations)
- impl - expected value is a full name of type with
package name `github.com/proj/implement/SomeStructName`.
It will be used in case you have registered multiple
implementations of one interface. So object with provided
name will be injected.
- qual - expected value is any string. It will be used
in case you have registered multiple implementations
of one interface. Implementation will be chosen based on
value provided in this config and value that will be
returned by `Qualify() string` method of one of
registered implementations.

***
Example

TODO: example for each of explained configuration

****
Any value by key

In case you want to inject not only implementation of interface, but
any kind of value you can use `reg` configuration of `fig` tag.
You should register all values you want to use with
`RegisterValue(key string, value interface{}) error` method.
So value in tag `fig:"reg[key1]"`, in our case it is `key1`, should
be registered with value that is assignable to type of field where
this value is expected.
- reg - expected value is any string. This key should be
registered with `RegisterValue` method.

***
Example

TODO: example for each of explained configuration

***
Environment variables injection

Scenario when you need some environment variable value is quite
popular, so there is solution for it. Tag configuration `env` solves
problem with manual initialization of such values.
- env - expected value is any string that can represent key of
environment variable. Value of this variable will be assigned to field

***
Example

TODO: example for each of explained configuration
