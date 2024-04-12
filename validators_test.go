package smapper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type eSrc struct {
	Int    int
	String string
}

type eDst struct {
	AnotherInt int `smapper:"Int,required"`
	String     string
}

type lSrc struct {
	String string
	Slice  []int
	Map    map[int]int
	Chan   chan struct{}
}

type lDst struct {
	String string        `smapper:",len=14"`
	Slice  []int         `smapper:",len=4"`
	Map    map[int]int   `smapper:",len=3"`
	Chan   chan struct{} `smapper:",len=2"`
}

type cSrc struct {
	Int     int
	Uint    uint
	Float   float64
	Complex complex128
	String  string
	Slice   []int
	Map     map[int]int
	Chan    chan struct{}
}

type cDst struct {
	Int     int           `smapper:",gte=-5,lte=-3,gt=-5,lt=-3,eq=-4,ne=-1,required"`
	Uint    uint          `smapper:",gte=5,lte=7,gt=5,lt=8,eq=6,ne=4,required"`
	Float   float64       `smapper:",gte=3.14,lte=4,gt=3,lt=4,eq=3.14,ne=1.9532,required"`
	Complex complex128    `smapper:",eq=4,ne=2,required"`
	String  string        `smapper:",len=14,gte=14,lte=14,gt=13,lt=15,eq=this is a test,ne=test,required"`
	Slice   []int         `smapper:",len=4,gte=3,lte=5,gt=3,lt=5,eq=4,ne=2,required"`
	Map     map[int]int   `smapper:",len=3,gte=3,lte=4,gt=2,lt=4,eq=3,ne=1,required"`
	Chan    chan struct{} `smapper:",len=2,gte=2,lte=2,gt=1,lt=3,eq=2,ne=1,required"`
}

type uElem struct {
	Int int
}

type uSrc struct {
	Slice []uElem
	Map   map[int]uElem
}

type uDst struct {
	Slice []uElem       `smapper:",unique"`
	Map   map[int]uElem `smapper:",unique"`
}

func Test_exists(t *testing.T) {
	t.Parallel()

	mapper := New()

	exist := eSrc{
		Int:    1,
		String: "simple",
	}
	notExist := eSrc{String: "simple1"}
	dst := &eDst{}

	assert.Nil(t, mapper.Map(exist, dst), "should be nil because Int exists and have a value")
	assert.Error(t, mapper.Map(notExist, dst), "should have errors because Int (required value) does not exists")
}

func Test_isUnique(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		src       uSrc
		expectErr bool
	}

	tests := []testCase{
		{
			"all good",
			uSrc{Slice: []uElem{{1}, {2}, {3}}, Map: map[int]uElem{1: {1}, 2: {2}, 3: {3}}},
			false,
		},
		{
			"Slice is not unique",
			uSrc{Slice: []uElem{{1}, {1}, {1}}, Map: map[int]uElem{1: {1}, 2: {2}, 3: {3}}},
			true,
		},
		{
			"Map is not unique",
			uSrc{Slice: []uElem{{1}, {2}, {3}}, Map: map[int]uElem{1: {1}, 2: {1}, 3: {1}}},
			true,
		},
		{
			"both Slice and Map are not unique",
			uSrc{Slice: []uElem{{2}, {2}, {3}}, Map: map[int]uElem{1: {1}, 2: {1}, 3: {3}}},
			true,
		},
		{
			"Slice is empty",
			uSrc{Slice: []uElem{}, Map: map[int]uElem{1: {1}, 2: {2}, 3: {3}}},
			false,
		},
		{
			"Map is empty",
			uSrc{Slice: []uElem{{1}, {2}, {3}}, Map: map[int]uElem{}},
			false,
		},
		{
			"Slice is nil",
			uSrc{Slice: nil, Map: map[int]uElem{1: {1}, 2: {2}, 3: {3}}},
			false,
		},
		{
			"Map is nil",
			uSrc{Slice: []uElem{{1}, {2}, {3}}, Map: nil},
			false,
		},
	}

	mapper := New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dst := &uDst{}

			if test.expectErr {
				assert.Error(t, mapper.Map(test.src, dst))
			} else {
				assert.Nil(t, mapper.Map(test.src, dst))
			}
		})
	}
}

func Test_hasLen(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		value     lSrc
		expectErr bool
	}

	tests := []testCase{
		{
			"all good",
			lSrc{"this is a test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(2)},
			false,
		},
		{
			"String's len is not equal",
			lSrc{"this is not a test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(2)},
			true,
		},
		{
			"Slice's len is not equal",
			lSrc{"this is a test", []int{1, 2, 3}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(2)},
			true,
		},
		{
			"Map's len is not equal",
			lSrc{"this is a test", []int{1, 2, 3, 4}, map[int]int{1: 1}, makeChan(2)},
			true,
		},
		{
			"Chan's len is not equal",
			lSrc{"this is a test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(3)},
			true,
		},
	}

	mapper := New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := &lDst{}

			if test.expectErr {
				assert.Error(t, mapper.Map(test.value, d))
			} else {
				assert.Nil(t, mapper.Map(test.value, d))
			}
		})
	}
}

func Test_compare(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		value     cSrc
		expectErr bool
	}

	mapper := New()

	tests := []testCase{
		{
			"all good",
			cSrc{-4, 6, 3.14, complex(4, 4), "this is a test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(2)},
			false,
		},
		{
			"Int is not equal",
			cSrc{-5, 6, 3.14, complex(4, 4), "this is a test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(2)},
			true,
		},
		{
			"Uint is not equal",
			cSrc{-4, 8, 3.14, complex(4, 4), "this is a test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(2)},
			true,
		},
		{
			"Float is not equal",
			cSrc{-4, 6, 3.144, complex(4, 4), "this is a test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(2)},
			true,
		},
		{
			"Complex is not equal",
			cSrc{-4, 6, 3.14, complex(4, 6), "this is a test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(2)},
			true,
		},
		{
			"String is not equal",
			cSrc{-4, 6, 3.14, complex(4, 4), "test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(2)},
			true,
		},
		{
			"Slice is not equal",
			cSrc{-4, 6, 3.14, complex(4, 4), "this is a test", []int{1, 2, 3, 4, 5}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(2)},
			true,
		},
		{
			"Map is not equal",
			cSrc{-4, 6, 3.14, 4, "this is a test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2}, makeChan(2)},
			true,
		},
		{
			"Chan is not equal",
			cSrc{-4, 6, 3.14, 4, "this is a test", []int{1, 2, 3, 4}, map[int]int{1: 1, 2: 2, 3: 3}, makeChan(4)},
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dst := &cDst{}

			if test.expectErr {
				assert.Error(t, mapper.Map(test.value, dst))
			} else {
				assert.Nil(t, mapper.Map(test.value, dst))
			}
		})
	}
}

func makeChan(n int) chan struct{} {
	ch := make(chan struct{}, n)

	for i := 0; i < n; i++ {
		ch <- struct{}{}
	}

	return ch
}
