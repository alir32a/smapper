package smapper

import (
	"cmp"
	"fmt"
	"reflect"
	"strconv"
)

type validator struct {
	name  string
	param string
	fn    ValidatorFunc
}

var defaultValidators = map[string]ValidatorFunc{
	"required": exists,
	"unique":   isUnique,
	"len":      hasLen,
	"gte":      hasGte,
	"gt":       hasGt,
	"lte":      hasLte,
	"lt":       hasLt,
	"eq":       equals,
	"ne":       notEquals,
}

func exists(v reflect.Value, param string) bool {
	return !v.IsZero()
}

func isUnique(v reflect.Value, param string) bool {
	value := reflect.ValueOf(struct{}{})

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		el := v.Type().Elem()
		if el.Kind() == reflect.Ptr {
			el = el.Elem()
		}

		set := reflect.MakeMapWithSize(reflect.MapOf(el, value.Type()), v.Len())
		for i := 0; i < v.Len(); i++ {
			set.SetMapIndex(reflect.Indirect(v.Index(i)), value)
		}

		return set.Len() == v.Len()
	case reflect.Map:
		iter := v.MapRange()

		keyType := v.Type().Elem()
		if keyType.Kind() == reflect.Ptr {
			keyType = keyType.Elem()
		}

		set := reflect.MakeMapWithSize(reflect.MapOf(keyType, value.Type()), v.Len())
		for iter.Next() {
			val := iter.Value()
			set.SetMapIndex(reflect.Indirect(val), value)
		}

		return set.Len() == v.Len()
	default:
		panic(fmt.Sprintf("unsupported type for unique, want map, slice or array got %s", v.Type().Name()))
	}
}

func hasLen(v reflect.Value, param string) bool {
	n, err := strconv.Atoi(param)
	if err != nil {
		panic(fmt.Sprintf("%s is not a valid number", param))
	}
	l := v.Len()
	return l == n
}

func compare(v reflect.Value, param string) int {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("invalid param, %s", err.Error()))
		}

		return cmp.Compare(v.Int(), n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("invalid param, %s", err.Error()))
		}

		return cmp.Compare(v.Uint(), n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(param, 64)
		if err != nil {
			panic(fmt.Sprintf("invalid param, %s", err.Error()))
		}

		return cmp.Compare(v.Float(), n)
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan, reflect.String:
		n, err := strconv.Atoi(param)
		if err != nil {
			panic(fmt.Sprintf("invalid param, %s", err.Error()))
		}

		return cmp.Compare(v.Len(), n)
	default:
		panic(fmt.Sprintf("unsupported type for compare: %s", v.Type().Name()))
	}
}

func hasGte(v reflect.Value, param string) bool {
	return compare(v, param) >= 0
}

func hasGt(v reflect.Value, param string) bool {
	return compare(v, param) > 0
}

func hasLte(v reflect.Value, param string) bool {
	return compare(v, param) <= 0
}

func hasLt(v reflect.Value, param string) bool {
	return compare(v, param) < 0
}

func equals(v reflect.Value, param string) bool {
	switch v.Kind() {
	// since complex numbers have no ordering, we can just compare them for equality
	case reflect.Complex64, reflect.Complex128:
		n, err := strconv.ParseFloat(param, 64)
		if err != nil {
			panic(fmt.Sprintf("invalid param, %s", err.Error()))
		}

		return isComplexesEqual(v.Complex(), n)
	case reflect.String:
		return cmp.Compare(v.String(), param) == 0
	}

	return compare(v, param) == 0
}

func notEquals(v reflect.Value, param string) bool {
	return !equals(v, param)
}

func isComplexesEqual(c complex128, n float64) bool {
	a := cmp.Compare(real(c), n)
	b := cmp.Compare(imag(c), n)
	_, _ = a, b
	return cmp.Compare(real(c), n) == 0 && cmp.Compare(imag(c), n) == 0
}
