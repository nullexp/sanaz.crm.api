package utility

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type person struct {
	Name     string
	LastName string
}

type address struct {
	Country string
	City    string
}

type client struct {
	Name     string
	LastName string
	Address  address
}

type clientNullableAddress struct {
	Name     string
	LastName string
	Address  *address
}

type testCase[T comparable] struct {
	name    string
	current T
	new     T
	err     error
	cmp     map[string]any
}

func TestCompareOnPointerOfPerson(t *testing.T) {
	for _, testCase := range testCaseOnPersonPointer {
		t.Run(testCase.name, func(t *testing.T) {
			cmp, err := Compare[*person](testCase.current, testCase.new)
			assert.EqualValues(t, err, testCase.err)
			assert.EqualValues(t, cmp, testCase.cmp)
		})
	}
}

func TestCompareOnPerson(t *testing.T) {
	for _, testCase := range testCaseOnPerson {
		t.Run(testCase.name, func(t *testing.T) {
			cmp, err := Compare[person](testCase.current, testCase.new)
			assert.EqualValues(t, err, testCase.err)
			assert.EqualValues(t, cmp, testCase.cmp)
		})
	}
}

//func TestCompareOnPointerOfClient(t *testing.T) {
//	for _, testCase := range testCaseOnClient {
//		t.Run(testCase.name, func(t *testing.T) {
//			cmp, err := Compare[*client](testCase.current, testCase.new)
//			assert.EqualValues(t, err, testCase.err)
//			assert.EqualValues(t, cmp, testCase.cmp)
//		})
//	}
//}

func TestCompareOnClient(t *testing.T) {
	for _, testCase := range testCaseOnClient {
		t.Run(testCase.name, func(t *testing.T) {
			cmp, err := Compare[client](testCase.current, testCase.new)
			assert.EqualValues(t, err, testCase.err)
			assert.EqualValues(t, cmp, testCase.cmp)
		})
	}
}

func TestNestedNullableEntity(t *testing.T) {
	for _, testCase := range testCaseOnClientWithNullableAddress {
		t.Run(testCase.name, func(t *testing.T) {
			cmp, err := Compare[clientNullableAddress](testCase.current, testCase.new)
			assert.EqualValues(t, err, testCase.err)
			assert.EqualValues(t, cmp, testCase.cmp)
		})
	}
}

var testCaseOnPersonPointer = []testCase[*person]{
	{
		name: "Error on the current argument",
		new: &person{
			Name:     "Nima",
			LastName: "Zare",
		},
		err: CompareErrorCurrentCannotBeNil,
	},
	{
		name: "Error on the new argument",
		current: &person{
			Name:     "Nima",
			LastName: "Zare",
		},
		err: CompareErrorNewCannotBeNil,
	},
	{
		name: "No different",
		current: &person{
			Name:     "Nima",
			LastName: "Zare",
		},
		new: &person{
			Name:     "Nima",
			LastName: "Zare",
		},
	},
	{
		name: "Name isChanged",
		current: &person{
			Name:     "Nima",
			LastName: "Zare",
		},
		new: &person{
			Name:     "Nimaa",
			LastName: "Zare",
		},
		cmp: map[string]any{
			"Name": "Nimaa",
		},
	},
	{
		name: "LastName isChanged",
		current: &person{
			Name:     "Nima",
			LastName: "Zareh",
		},
		new: &person{
			Name:     "Nima",
			LastName: "Zare",
		},
		cmp: map[string]any{
			"LastName": "Zare",
		},
	},
}

var testCaseOnPerson = []testCase[person]{
	{
		name: "Empty current, returm whole new value",
		new: person{
			Name:     "Nima",
			LastName: "Zare",
		},
		cmp: map[string]any{
			"Name":     "Nima",
			"LastName": "Zare",
		},
	},
	{
		name: "Enpty entity as new, cause empty fileds",
		current: person{
			Name:     "Nima",
			LastName: "Zare",
		},
		cmp: map[string]any{
			"Name":     "",
			"LastName": "",
		},
	},
	{
		name: "No different",
		current: person{
			Name:     "Nima",
			LastName: "Zare",
		},
		new: person{
			Name:     "Nima",
			LastName: "Zare",
		},
	},
	{
		name: "Name isChanged",
		current: person{
			Name:     "Nima",
			LastName: "Zare",
		},
		new: person{
			Name:     "Nimaa",
			LastName: "Zare",
		},
		cmp: map[string]any{
			"Name": "Nimaa",
		},
	},
	{
		name: "LastName isChanged",
		current: person{
			Name:     "Nima",
			LastName: "Zareh",
		},
		new: person{
			Name:     "Nima",
			LastName: "Zare",
		},
		cmp: map[string]any{
			"LastName": "Zare",
		},
	},
}

var testCaseOnClient = []testCase[client]{
	{
		name: "Just city changes",
		current: client{
			Name:     "Nima",
			LastName: "Zare",
			Address: address{
				Country: "Iran",
				City:    "Tehran",
			},
		},
		new: client{
			Name:     "Nima",
			LastName: "Zare",
			Address: address{
				Country: "Iran",
				City:    "Shiraz",
			},
		},
		cmp: map[string]any{
			"Address": map[string]any{
				"City": "Shiraz",
			},
		},
	},
}

var testCaseOnClientWithNullableAddress = []testCase[clientNullableAddress]{
	{
		name: "Just city changes",
		current: clientNullableAddress{
			Name:     "Nima",
			LastName: "Zare",
			Address: &address{
				Country: "Iran",
				City:    "Tehran",
			},
		},
		new: clientNullableAddress{
			Name:     "Nima",
			LastName: "Zare",
			Address: &address{
				Country: "Iran",
				City:    "Shiraz",
			},
		},
		cmp: map[string]any{
			"Address": map[string]any{
				"City": "Shiraz",
			},
		},
	},
}
