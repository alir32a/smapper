package smapper

import "reflect"

type FieldValue struct {
	reflect.Value
	ParentType reflect.Type
	FieldName  string
}

func (f FieldValue) From(v reflect.Value) FieldValue {
	return FieldValue{
		Value:      v,
		ParentType: f.ParentType,
		FieldName:  f.FieldName,
	}
}

func NewFieldValue(value reflect.Value, parent reflect.Type, name string) FieldValue {
	return FieldValue{
		Value:      value,
		ParentType: parent,
		FieldName:  name,
	}
}

type CallbackFunc func(reflect.Type, reflect.Type, any) (any, error)

type ValidatorFunc func(reflect.Value, string) bool

type Validator struct {
	Name string
	Func ValidatorFunc
}

func NewValidator(name string, fn ValidatorFunc) *Validator {
	return &Validator{
		Name: name,
		Func: fn,
	}
}

type Callback struct {
	Name string
	Func CallbackFunc
}

func NewCallback(name string, fn CallbackFunc) *Callback {
	return &Callback{
		Name: name,
		Func: fn,
	}
}
