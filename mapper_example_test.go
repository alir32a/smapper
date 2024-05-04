package smapper_test

import (
	"errors"
	"fmt"
	"github.com/alir32a/smapper"
	"reflect"
	"strings"
)

func ExampleMapper_Map() {
	type Person struct {
		ID       uint
		Username string
	}

	type User struct {
		ID   int64
		Name string `smapper:"username"`
	}

	mapper := smapper.New()

	person := Person{
		ID:       42,
		Username: "alir32a",
	}
	user := User{}

	err := mapper.Map(person, &user)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.ID)   // 42
	fmt.Println(user.Name) // alir32a
}

func ExampleMapper_Map_validators() {
	type Person struct {
		ID          uint
		Username    string
		PhoneNumber string
	}

	type User struct {
		ID          int64  `smapper:",required"`
		Name        string `smapper:"username,eq=admin"`
		PhoneNumber string `smapper:",len=8"`
	}

	p := Person{
		ID:          42,         // should not be 0
		Username:    "admin",    // must be equal to "admin"
		PhoneNumber: "12345678", // must have exactly 8 characters
	}
	user := User{}

	err := smapper.Map(p, &user) // must be ok
	if err != nil {
		panic(err)
	}

	p.ID = 0
	p.Username = "common"
	p.PhoneNumber = "1234567"

	err = smapper.Map(p, &user) // returns validation error
	if err != nil {
		fmt.Println(err.Error())
	}
}

func ExampleMapper_Map_callbacks() {
	type Person struct {
		ID       uint
		Username string
	}

	type User struct {
		ID   int64
		Name string `smapper:"Username,callback:uppercase"`
	}

	person := Person{
		ID:       42,
		Username: "admin",
	}
	user := User{}

	mapper := smapper.New(smapper.WithCallbacks(
		smapper.NewCallback("uppercase", func(src reflect.Type, dst reflect.Type, v any) (any, error) {
			// you have access to source and destination types here
			if src.Kind() != reflect.String || dst.Kind() != reflect.String {
				return nil, errors.New("wrong type")
			}

			return strings.ToUpper(v.(string)), nil
		},
		)))

	err := mapper.Map(person, &user)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.Name) // ADMIN
}

func ExampleMapAndReturn() {
	type Person struct {
		ID       uint
		Username string
	}

	type User struct {
		ID   int64
		Name string `smapper:"username"`
	}

	person := Person{
		ID:       42,
		Username: "alir32a",
	}

	user, err := smapper.MapTo[User](person)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.ID)   // 42
	fmt.Println(user.Name) // alir32a
}
