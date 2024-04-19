# smapper

[![Go Report Card](https://goreportcard.com/badge/github.com/alir32a/smapper)](https://goreportcard.com/report/github.com/alir32a/smapper)
![tests workflow](https://github.com/alir32a/smapper/actions/workflows/tests.yml/badge.svg?branch=main)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/alir32a/smapper/checks.yml?label=build)
[![Go Reference](https://pkg.go.dev/badge/github.com/alir32a/smapper.svg)](https://pkg.go.dev/github.com/alir32a/smapper)

A flexible and easy-to-use struct-to-struct mapping library allows you to convert an arbitrary type into another arbitrary type. It
supports custom validations to ensure data consistency, automatically converts compatible types, and 
gives you control over incompatible types (e.g. string to time.Time) using callbacks you specify per each field.

## Installation

Using ```go get```

```
$ go get github.com/alir32a/smapper
```

## Usages and Examples

### Convert two similar structs

```go
package main

import (
	"fmt"
	"github.com/alir32a/smapper"
)

type User struct {
	ID       uint
	Username string
}

type Person struct {
	ID   int64 `smapper:"-"` // will be skipped
	UserID uint `smapper:"ID"`
	Name string `smapper:"username"`
}

func main() {
	mapper := smapper.New()

	user := User{
		ID:       42,
		Username: "alir32a",
	}
	person := Person{}

	err := mapper.Map(user, &person)
	if err != nil {
		panic(err)
	}

	fmt.Println(person.ID) // 0
	fmt.Println(person.UserID)   // 42
	fmt.Println(person.Name) // alir32a
}
```

### Validations

```go
package main

import (
	"fmt"
	"reflect"
	"regexp"
	"github.com/alir32a/smapper"
	"strings"
)

type User struct {
	ID       uint
	Username string
	Email    string
}

type Person struct {
	ID     uint   `smapper:"-"` // will be skipped
	UserID int64  `smapper:",required"`
	Name   string `smapper:"username,eq=admin"`
	Email  string `smapper:",contains=gmail"`
}

func main() {
	mapper := smapper.New(smapper.WithValidators(Contains()))
    
	user := User{
		ID: 42, // cannot be 0
		Name: "admin", // should be equal to admin
		Email: "admin@gmail.com", // must contain "gmail"
    }
	person := Person{}
	
	err := mapper.Map(user, &person)
	if err != nil {
        panic("should be nil")
	}
}

func Contains() *smapper.Validator {
	return smapper.NewValidator("contains", func(v reflect.Value, param string) bool {
		if v.Kind() != reflect.String {
			panic("expected string")
		}

		return strings.Contains(v.String(), param)
	})
}
```

You can define your own validators, just like we did here. 
We created a function to check if the value contains the required parameter ('gmail'). 
If it doesn't, validation fails, resulting in a validation error. 
You can use multiple validators on a single field, combining built-in validators with your custom ones.

#### Built-in Validators

| Tag Name | Description                                                                  | Accept Parameter |
|----------|------------------------------------------------------------------------------|----|
| required | Field must not be zero value                                                 | No
| unique | Field must have unique values<br/> (can be used with slices, arrays and maps) | No
| len | Field must have the given length                                             | Yes
| eq | Field must be equal to the given param                                       | Yes
| ne | Field must not be equal to the given param                                   | Yes
| gte | Field's value or length must be greater than or equal to the given param     | Yes
| gt | Field's value or length must be greater than the given param                 | Yes
| lte | Field's value or length must be less than or equal to the given param        | Yes
| lt | Field's value or length must be greater than the given param      | Yes

> To override built-in validators, set `Mapper.OverrideDefaultValidators = true` or 
> use `WithOverrideDefaultValidators()` during initialization.

### Callbacks

```go
package main

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"github.com/alir32a/smapper"
	"strconv"
)

type User struct {
	ID       int64
	Username string
	Email    string
}

type Person struct {
	ID     uint    `smapper:"-"`
	UserID string  `smapper:"ID,callback:to_string"`
	Name   string  `smapper:"username"`
	Email  string
}

func main() {
	mapper := smapper.New(smapper.WithCallbacks(ToString()))

	user := User{
		ID:       42,                // converts to "42"
		Username: "admin",
		Email:    "admin@gmail.com",
	}
	person := Person{}

	err := mapper.Map(user, &person)
	if err != nil {
		panic("should be nil")
	}
	
	fmt.Println(person.UserID) // 42
}

func ToString() *smapper.Callback {
	return smapper.NewCallback("to_string", func(src reflect.Type, target reflect.Type, v any) (any, error) {
		if target.Kind() != reflect.String {
			return v, errors.New("target type is not string")
		}

		switch src.Kind() {
		case reflect.Int64:
			return strconv.FormatInt(v.(int64), 10), nil
		case reflect.Uint64:
			return strconv.FormatUint(v.(uint64), 10), nil
		case reflect.Float64:
			return strconv.FormatFloat(v.(float64), 'f', -1, 64), nil
		case reflect.String:
			return v, nil
		default:
            return v, errors.New("unsupported type")
		}
	})
}
```

> you can execute only **one** callback per field.

we defined a callback named `to_string`, to map an `int64` field to `string`. because int64 and strings
are incompatible (you cannot use go type conversion like string(int64)), you need to define a callback to do this conversion.
as you can see, you have access to both input and output types in the callback function, so you can safely do the conversion.

> smapper can automatically convert between strings and numbers, but it's disabled by default, to enable it,
>  you need to set `AutoStringToNumberConversion = true` and/or `AutoNumberToStringConversion = true`, or use 
>  `WithAutoStringToNumberConversion()` and/or `WithAutoNumberToStringConversion()` when initializing a new mapper.

### Nested Structures

```go
package main

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"github.com/alir32a/smapper"
	"strconv"
	"time"
)
type PurchasedItem struct {
	ProductID uint
	Name      string
}

type ReturnedItem struct {
	ProductID uint
	Name      string
}

type Address struct {
	Title string
	Detail string
}

type Purchase struct {
	Items       []PurchasedItem
	Source      Address
	Destination Address
}

type Return struct {
	Items       []ReturnedItem
	Source      Address `smapper:"destination"`
	Destination Address `smapper:"source"`
}

func main() {
	mapper := smapper.New()

	home := Address{Title: "Home", Detail: "Pine St"}
	bakery := Address{Title: "Bakery", Detail: "East Ave"}
	pie := PurchasedItem{ProductID: 42, Name: "Apple pie"}

	purchase := Purchase{
		Items:       []PurchasedItem{pie},
		Source:      bakery,
		Destination: home,
	}
	ret := Return{}

	err := mapper.Map(purchase, &ret)
	if err != nil {
		panic("should be nil")
	}

	fmt.Printf("%+v\n", ret.Items)     // [{ProductID:42 Name:Apple pie}]
	fmt.Println(ret.Source.Title)      // Home
	fmt.Println(ret.Destination.Title) // Bakery
}
```

## What's Next?

- Implement more built-in validators.
- Support automatic mapping between different types of maps.
(e.g. map[int]User to map[string]Admin, currently you need to write a callback to the conversion)

I'd be happy to hear any suggestions or improvements you have.

## Contribution

Thanks for taking the time to contribute. Please see [CONTRIBUTING.md](https://github.com/alir32a/smapper/blob/main/CONTRIBUTING.md).

## Donations

Thank you for your encouraging support, It helps me build more things.

bitcoin: `bc1q44ffnc274fg0j982tjqthrn9ka85l8hhntgj82`
