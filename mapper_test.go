package smapper

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"math"
	"reflect"
	"strconv"
	"testing"
)

type Simple struct {
	Int    int
	String string
	float  float64
}

type AnotherSimple struct {
	Uint32 uint32 `smapper:"Int"`
	String string
	Float  float64
}

type Complex struct {
	Simple
	Slice     []Simple
	Map       map[int]Simple
	Simples   []Simple
	Complexes []Complex
}

type AnotherComplex struct {
	AnotherSimple AnotherSimple `smapper:"simple"`
	Slice         []Simple
	Map           map[float64]AnotherSimple
	Simples       []Simple `smapper:"-"`
	Complexes     []AnotherComplex
}

func TestMap_Simple(t *testing.T) {
	t.Parallel()

	mapper := New()

	first := Simple{
		Int:    1,
		String: "simple",
		float:  2.32,
	}
	second := &AnotherSimple{}

	err := mapper.Map(first, second)
	if err != nil {
		t.Error(err.Error())
	}

	assert.EqualValues(t, first.Int, second.Uint32, "first.Int should be equal to second.Uint32")
	assert.Equal(t, first.String, second.String, "first.String should be equal to second.String")
	assert.NotEqual(t, first.float, second.Float, "first.float is unexported and should not be equal to second.Float")
}

func TestMap_Complex(t *testing.T) {
	t.Parallel()

	mapper := New()

	simples := []Simple{
		{
			Int:    1,
			String: "simple1",
			float:  1.234,
		},
		{
			Int:    1,
			String: "simple2",
			float:  3.434,
		},
		{
			Int:    1,
			String: "simple3",
			float:  5.734,
		},
	}
	src := Complex{
		Simple: Simple{
			Int:    1,
			String: "simple",
			float:  1.23,
		},
		Slice: simples,
		Map: map[int]Simple{
			42:  simples[0],
			17:  simples[1],
			154: simples[2],
		},
		Simples: simples,
		Complexes: []Complex{
			{
				Simple: Simple{
					Int:    1,
					String: "simple",
					float:  1.23,
				},
			},
		},
	}
	dst := &AnotherComplex{}

	assert.NoError(t, mapper.Map(src, dst))

	assert.EqualValues(t, src.Simple.Int, dst.AnotherSimple.Uint32,
		"src.Simple.Int should be equal to dst.AnotherSimple.Uint32")
	assert.Equal(t, src.Simple.String, dst.AnotherSimple.String,
		"src.Simple.String should be equal to dst.AnotherSimple.String")
	assert.NotEqual(t, src.Simple.float, dst.AnotherSimple.Float,
		"src.Simple.float is unexported and should not be equal to dst.AnotherSimple.Float")

	for k, v := range src.Map {
		dstVal, found := dst.Map[float64(k)]
		assert.Truef(t, found, "key %d not found on destination", k)

		assert.EqualValues(t, v.Int, dstVal.Uint32)
		assert.Equal(t, v.String, dstVal.String)
		assert.NotEqual(t, v.float, dstVal.Float)
	}

	assert.Equal(t, src.Slice, dst.Slice, "src.Slice should be equal to dst.Slice")
	assert.Nil(t, dst.Simples, "dst.Simples have an ignore tag, so it should be nil")

	assert.EqualValues(t, src.Complexes[0].Simple.Int, dst.Complexes[0].AnotherSimple.Uint32,
		"src.Complexes[0].Simple.Int should be equal to dst.Complexes[0].AnotherSimple.Uint32")
	assert.Equal(t, src.Complexes[0].Simple.String, dst.Complexes[0].AnotherSimple.String,
		"src.Complexes[0].Simple.String should be equal to dst.Complexes[0].AnotherSimple.String")
	assert.NotEqual(t, src.Complexes[0].Simple.float, dst.Complexes[0].AnotherSimple.Float,
		"src.Complexes[0].Simple.float is unexported and should not be equal to dst.Complexes[0].AnotherSimple.Float")
}

func TestMap_WrongTypes(t *testing.T) {
	t.Parallel()

	mapper := New()

	src := Simple{
		Int:    1,
		String: "simple",
		float:  1.345,
	}
	dst := Simple{}

	// test non-pointer dst
	assert.Error(t, mapper.Map(src, dst), "should have error because dst is not a pointer")

	// test non-struct src and dst
	notStruct := []Simple{src}
	assert.Error(t, mapper.Map(src, &notStruct), "should have error because dst is a pointer to list")
	assert.Error(t, mapper.Map(notStruct, &dst), "should have error because src is not a struct")

	// test true inputs
	assert.NoError(t, mapper.Map(src, &dst))
}

func TestMap_DifferentTypes(t *testing.T) {
	t.Parallel()

	mapper := New(WithAutoStringToNumberConversion())

	type src struct {
		Int    int
		String string
		Float  float64
		Slice  []int
	}

	type dst struct {
		Uint32 uint32 `smapper:"Int"`
		StrInt int64  `smapper:"String"`
		Int    int    `smapper:"Float"`
		Slice  []float64
	}

	s := src{
		Int:    12,
		String: "1234",
		Float:  3.14,
		Slice:  []int{1, 2, 3, 4},
	}
	d := dst{}

	assert.NoError(t, mapper.Map(s, &d))

	assert.EqualValues(t, s.Int, d.Uint32)
	assert.Equal(t, strconv.FormatInt(d.StrInt, 10), s.String)
	assert.EqualValues(t, math.Floor(s.Float), d.Int)
	assert.Equal(t, len(s.Slice), len(d.Slice))

	for i := range d.Slice {
		assert.EqualValues(t, s.Slice[i], d.Slice[i])
	}
}

func TestMap_Callback(t *testing.T) {
	t.Parallel()

	type simple struct {
		I int `smapper:",callback:double"`
	}

	withCallback := New(WithCallbacks(NewCallback("double", func(src, dst reflect.Type, v any) (any, error) {
		u, ok := v.(int)
		if !ok {
			return v, errors.New("v should be uint")
		}

		return u * 2, nil
	})))
	withoutCallback := New()

	src := simple{I: 2}
	dst := &simple{}

	assert.Nil(t, withCallback.Map(src, dst))
	assert.Equal(t, dst.I, 4, "dst.Uint should be 2*2=4")

	assert.Error(t, withoutCallback.Map(src, dst), "should have error because 'double' callback does not exist")
}

func TestMap_CustomValidators(t *testing.T) {
	t.Parallel()

	type simple struct {
		I int `smapper:",is_even"`
	}

	withValidator := New(WithValidators(NewValidator("is_even", func(v reflect.Value, s string) bool {
		u := v.Int()

		return u%2 == 0
	})))
	withoutValidator := New()

	even := simple{I: 2}
	notEven := simple{I: 1}
	dst := &simple{}

	assert.Error(t, withValidator.Map(notEven, dst), "should have error because 1 is not even (is_even validator)")
	assert.Nil(t, withValidator.Map(even, dst))

	assert.Error(t, withoutValidator.Map(even, dst), "should have error because 'is_even' validator does not exist")
}

func TestMap_OverrideDefaultValidators(t *testing.T) {
	t.Parallel()

	type simple struct {
		I int `smapper:",eq=2"`
	}

	opt := WithValidators(NewValidator("eq", func(v reflect.Value, param string) bool {
		u := v.Int()
		p, _ := strconv.ParseInt(param, 10, 64)

		return u != p
	}))

	with := New(opt, WithOverrideDefaultValidators())
	without := New(opt)

	src := simple{I: 2}
	dst := &simple{}

	assert.Error(t, with.Map(src, dst),
		"should have error because we overriden eq func to do the opposite (OverrideDefaultValidators=true)")
	assert.Nil(t, without.Map(src, dst), "should be nil, OverrideDefaultValidators=false")
}

func TestMap_IgnoreMissingValidators(t *testing.T) {
	t.Parallel()

	type simple struct {
		Int int `smapper:",invalid_validator=2"`
	}

	with := New(WithIgnoreMissingValidators())
	without := New()

	src := &simple{1}
	dst := &simple{}

	assert.Nil(t, with.Map(src, dst), "should not have error because IgnoreMissingValidators = true")
	assert.Error(t, without.Map(src, dst),
		"should have error because invalid_validator does not exist and IgnoreMissingValidators = false")
}

func TestMap_IgnoreMissingCallbacks(t *testing.T) {
	t.Parallel()

	type simple struct {
		Int int `smapper:",callback:invalid_callback"`
	}

	with := New(WithIgnoreMissingCallbacks())
	without := New()

	src := &simple{1}
	dst := &simple{}

	assert.Nil(t, with.Map(src, dst), "should not have error because IgnoreMissingCallbacks = true")
	assert.Error(t, without.Map(src, dst),
		"should have error because invalid_callback does not exist and IgnoreMissingCallbacks = false")
}

func TestMap_AutoNumberToStringConversion(t *testing.T) {
	t.Parallel()

	validMapper := New(
		WithAutoStringToNumberConversion(),
		WithAutoNumberToStringConversion(),
	)
	invalidMapper := New()

	type src struct {
		Int    int
		String string
	}

	type dst struct {
		StrInt string  `smapper:"int"`
		Float  float64 `smapper:"string"`
	}

	validSrc := src{
		Int:    123,
		String: "3.14",
	}
	invalidSrc := src{
		Int:    123,
		String: "not a number",
	}
	d := dst{}

	assert.Error(t, invalidMapper.Map(validSrc, &d))
	assert.Error(t, invalidMapper.Map(invalidSrc, &d))

	assert.Error(t, validMapper.Map(invalidSrc, &d))
	assert.NoError(t, validMapper.Map(validSrc, &d))

	assert.Equal(t, strconv.Itoa(validSrc.Int), d.StrInt)

	f, _ := strconv.ParseFloat(validSrc.String, 64)
	assert.Equal(t, f, d.Float)
}

func TestMapAndReturn(t *testing.T) {
	t.Parallel()

	simple := Simple{
		Int:    1,
		String: "simple",
		float:  2.32,
	}

	anotherSimple, err := MapTo[AnotherSimple](simple)
	assert.NoError(t, err)

	assert.EqualValues(t, simple.Int, anotherSimple.Uint32)
	assert.Equal(t, simple.String, anotherSimple.String)
	assert.NotEqual(t, simple.float, anotherSimple.Float)
}
