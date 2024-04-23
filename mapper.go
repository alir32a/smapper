package smapper

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	emptyTag    = ""
	ignoreTag   = "-"
	callbackTag = "callback:"
)

type Mapper struct {
	Config
	callbacks  map[string]CallbackFunc
	validators map[string]ValidatorFunc
}

// New returns a new Mapper with the given options.
func New(opts ...Option) *Mapper {
	mapper := &Mapper{
		callbacks:  make(map[string]CallbackFunc),
		validators: make(map[string]ValidatorFunc),
	}

	for _, opt := range opts {
		opt(mapper)
	}

	return mapper
}

// Map takes a struct and converts it into another struct, output type must be a pointer to a struct.
func (m *Mapper) Map(input, output any) error {
	err := validateInputTypes(reflect.TypeOf(input), reflect.TypeOf(output))
	if err != nil {
		return err
	}

	srcVal := reflect.ValueOf(input)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}
	dstVal := reflect.ValueOf(output).Elem()

	err = m.mapTypes(FieldValue{Value: srcVal}, FieldValue{Value: dstVal})
	if err != nil {
		return err
	}

	return nil
}

func (m *Mapper) mapTypes(src, dst FieldValue) error {
	for i := 0; i < dst.NumField(); i++ {
		dstField := dst.Field(i)

		// ignores the unexported field
		if !dst.Field(i).CanSet() {
			continue
		}

		field := dst.Type().Field(i)
		fieldName := field.Name

		// get the field name, callback function, and validator functions that
		// need to be executed before setting the value
		dstTags, err := m.parseTagValues(getTagValues(dst, field.Name))
		if err != nil {
			return err
		}

		if dstTags.field != emptyTag {
			if dstTags.field == ignoreTag {
				continue
			}

			// use the provided field name in the field tag instead of the actual field name
			fieldName = dstTags.field
		}

		// search for the field in the input type, and ignore it if it's zero value, or it's unexported
		if value := src.FieldByName(fieldName); value.IsValid() && value.CanInterface() {
			// execute parsed validators
			for _, v := range dstTags.validators {
				if !v.fn(value, v.param) {
					return &ValidationError{
						value:         NewFieldValue(value, src.Type(), field.Name),
						validatorName: v.name,
					}
				}
			}

			if dstTags.callback != nil {
				// execute parsed callback
				m, err := dstTags.callback(value.Type(), dstField.Type(), value.Interface())
				if err != nil {
					return &CallbackError{
						value: NewFieldValue(value, src.Type(), field.Name),
						msg:   err.Error(),
					}
				}

				value = reflect.ValueOf(m)
			}

			if value.Type() != dstField.Type() {
				// try to convert the source type to the destination type or return an error
				// if the conversion is impossible.
				v, err := m.convert(NewFieldValue(value, src.Type(), fieldName), dst.From(dstField))
				if err != nil {
					return err
				}

				value = v.Value
			}

			dstField.Set(value)
		}
	}

	return nil
}

// convert converts src type to dst type, returns error if the conversion is impossible. (e.g. map to slice).
func (m *Mapper) convert(src, dst FieldValue) (FieldValue, error) {
	if src.Type() == dst.Type() {
		return src, nil
	}

	switch dst.Type().Kind() {
	case reflect.Map:
		return src, &Error{msg: "mapping different types of maps doesn't supported"}
	case reflect.Slice, reflect.Array:
		if src.Type().Kind() != reflect.Slice && src.Type().Kind() != reflect.Array {
			return dst, &FieldError{
				value: src,
				msg:   fmt.Sprintf("cannot auto convert %s to %s", src.Type(), dst.Type()),
			}
		}

		dst.Grow(src.Len())
		dst.SetLen(src.Len())

		for i := 0; i < src.Len(); i++ {
			_, err := m.convert(src.From(src.Index(i)), dst.From(dst.Index(i)))
			if err != nil {
				return src, err
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err := m.convertInts(src, dst)
		if err != nil {
			return dst, err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		err := m.convertUints(src, dst)
		if err != nil {
			return dst, err
		}
	case reflect.Float32, reflect.Float64:
		err := m.convertFloats(src, dst)
		if err != nil {
			return dst, err
		}
	case reflect.String:
		err := m.convertStrings(src, dst)
		if err != nil {
			return dst, err
		}
	case reflect.Struct:
		err := m.mapTypes(src, dst)
		if err != nil {
			return src, err
		}
	default:
		return dst, &FieldError{
			value: dst,
			msg:   fmt.Sprintf("%s is not convertible", dst.Type()),
		}
	}

	return dst, nil
}

// convertInts converts uints, floats and strings to int.
func (m *Mapper) convertInts(src, dst FieldValue) error {
	src.Value = reflect.Indirect(src.Value)

	switch src.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dst.SetInt(src.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		dst.SetInt(int64(src.Uint()))
	case reflect.Float32, reflect.Float64:
		dst.SetInt(int64(src.Float()))
	case reflect.String:
		if !m.AutoStringToNumberConversion {
			return &FieldError{
				value: src,
				msg: fmt.Sprintf(
					"want %s, got string (if you want to auto convert strings to numbers, set AutoStringToNumberConversion to true",
					src.Type()),
			}
		}

		i, err := strconv.ParseInt(src.String(), 10, 64)
		if err != nil {
			return &Error{msg: "failed to auto convert"}
		}
		dst.SetInt(i)
	default:
		return &FieldError{
			value: src,
			msg:   fmt.Sprintf("cannot auto convert %s to %s", src.Type(), dst.Type()),
		}
	}

	return nil
}

// convertUints converts ints, floats and strings to uint.
func (m *Mapper) convertUints(src, dst FieldValue) error {
	src.Value = reflect.Indirect(src.Value)

	switch src.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dst.SetUint(uint64(src.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		dst.SetUint(src.Uint())
	case reflect.Float32, reflect.Float64:
		dst.SetUint(uint64(src.Float()))
	case reflect.String:
		if !m.AutoStringToNumberConversion {
			return &FieldError{
				value: src,
				msg: fmt.Sprintf(
					"want %s, got string (if you want to auto convert strings to numbers, set AutoStringToNumberConversion to true",
					src.Type()),
			}
		}

		i, err := strconv.ParseUint(src.String(), 10, 64)
		if err != nil {
			return &Error{msg: "failed to auto convert"}
		}
		dst.SetUint(i)
	default:
		return &FieldError{
			value: src,
			msg:   fmt.Sprintf("cannot auto convert %s to %s", src.Type(), dst.Type()),
		}
	}

	return nil
}

// convertFloats converts ints, uints and strings to float.
func (m *Mapper) convertFloats(src, dst FieldValue) error {
	src.Value = reflect.Indirect(src.Value)

	switch src.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dst.SetFloat(float64(src.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		dst.SetFloat(float64(src.Uint()))
	case reflect.Float32, reflect.Float64:
		dst.SetFloat(src.Float())
	case reflect.String:
		if !m.AutoStringToNumberConversion {
			return &FieldError{
				value: src,
				msg: fmt.Sprintf(
					"want %s, got string (if you want to auto convert strings to numbers, set AutoStringToNumberConversion to true",
					src.Type()),
			}
		}

		i, err := strconv.ParseFloat(src.String(), 64)
		if err != nil {
			return &Error{msg: "failed to auto convert"}
		}
		dst.SetFloat(i)
	default:
		return &FieldError{
			value: src,
			msg:   fmt.Sprintf("cannot auto convert %s to %s", src.Type(), dst.Type()),
		}
	}

	return nil
}

// convertStrings converts ints, uints and floats to string.
func (m *Mapper) convertStrings(src, dst FieldValue) error {
	src.Value = reflect.Indirect(src.Value)

	if src.Kind() != reflect.String && !m.AutoStringToNumberConversion {
		return &FieldError{
			value: src,
			msg: fmt.Sprintf(
				"want %s, got %s (if you want to auto convert numbers to strings, set AutoStringToNumberConversion to true",
				src.Type(), dst.Type()),
		}
	}

	switch src.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dst.SetString(strconv.FormatInt(src.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		dst.SetString(strconv.FormatUint(src.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		// using arguments that fmt.Println is already using to print floats.
		dst.SetString(strconv.FormatFloat(src.Float(), 'g', -1, 64))
	case reflect.String:
		dst.SetString(src.String())
	default:
		return &FieldError{
			value: src,
			msg:   fmt.Sprintf("cannot auto convert %s to %s", src.Type(), dst.Type()),
		}
	}

	return nil
}

// Map initialize a new Mapper with the given options and executes the Mapper.Map.
func Map(input, output any, opts ...Option) error {
	mapper := New(opts...)

	return mapper.Map(input, output)
}

type fieldOptions struct {
	field      string
	callback   CallbackFunc
	validators []validator
}

func (m *Mapper) parseTagValues(tags []string) (fieldOptions, error) {
	var res fieldOptions

	for i, tag := range tags {
		if i == 0 {
			res.field = toPascalCase(tag)

			continue
		}

		if funcName, found := strings.CutPrefix(tag, callbackTag); found {
			fn, found := m.callbacks[funcName]
			if !found {
				if m.IgnoreMissingCallbacks {
					continue
				}

				return fieldOptions{}, &Error{msg: fmt.Sprintf("cannot find callback %s", funcName)}
			}

			res.callback = fn
			continue
		}

		v, err := m.parseValidator(tag)
		if err != nil {
			if m.IgnoreMissingValidators {
				continue
			}

			return fieldOptions{}, err
		}

		res.validators = append(res.validators, v)
	}

	return res, nil
}

func (m *Mapper) parseValidator(tag string) (validator, error) {
	v := parseValidatorTag(tag)

	v.fn = defaultValidators[v.name]

	if m.validators != nil {
		if fn, found := m.validators[v.name]; found {
			if v.fn == nil || (v.fn != nil && m.OverrideDefaultValidators) {
				v.fn = fn
			}
		}
	}

	if v.fn == nil {
		return validator{}, &Error{msg: fmt.Sprintf("cannot find validator %s", v.name)}
	}

	return v, nil
}

func parseValidatorTag(tag string) validator {
	var v validator

	p := strings.Split(tag, "=")
	v.name = p[0]
	if len(p) > 1 {
		v.param = p[1]
	}

	return v
}

func getTagValues(v FieldValue, name string) []string {
	f, _ := v.Type().FieldByName(name)

	return strings.Split(f.Tag.Get("smapper"), ",")
}

func validateInputTypes(src, dst reflect.Type) error {
	if src.Kind() == reflect.Ptr {
		src = src.Elem()
	}

	if src.Kind() != reflect.Struct {
		return &Error{msg: fmt.Sprintf("input must be a struct or a pointer to struct, not %s", src.Kind())}
	}

	if dst.Kind() != reflect.Ptr || dst.Elem().Kind() != reflect.Struct {
		return &Error{msg: fmt.Sprintf("output must be a pointer to struct, not %s", dst.Kind())}
	}

	return nil
}

func toPascalCase(s string) string {
	if len(s) == 0 {
		return s
	}

	chars := strings.Split(s, "")
	chars[0] = strings.ToUpper(chars[0])

	return strings.Join(chars, "")
}
