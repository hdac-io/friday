package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodingNormalOperation1(t *testing.T) {
	fmt.Println("Test1: Basic situiation")

	var testName Name
	testName.Init("bryanrhee")

	stringName, _ := testName.ToString()
	fmt.Printf("Correct answer: 'bryanrhee'\nAnswer: '%s'\n", stringName)
	assert.EqualValues(t, stringName, "bryanrhee")
}

func TestEncodingNormalOperation2(t *testing.T) {
	fmt.Println("Test2: Basic situiation 2")

	var testName Name
	testName.Init("psy2848048")

	stringName, _ := testName.ToString()
	fmt.Printf("Correct answer: 'psy2848048'\nAnswer: '%s'\n", stringName)
	assert.EqualValues(t, stringName, "psy2848048")
}

func TestEncodingLongname(t *testing.T) {
	fmt.Println("Test3: Long name converting test")

	var testName Name
	testName.Init("psy.284_8048-i386")

	stringName, _ := testName.ToString()
	fmt.Printf("Correct answer: 'psy.284_8048-i386'\nAnswer: '%s'\n", stringName)
	assert.EqualValues(t, stringName, "psy.284_8048-i386")

}

func TestMapKeyAsStrct(t *testing.T) {
	fmt.Println("Test4: Feasiblility test of custom struct as a map key")

	var testName1 Name
	testName1.Init("psy.284_8048-i386")
	var testName2 Name
	testName2.Init("psy2848048")

	customMap := make(map[Name]string)
	customMap[testName1] = "Hello"
	customMap[testName2] = "World"

	assert.EqualValues(t, customMap[testName1], "Hello")
	assert.EqualValues(t, customMap[testName2], "World")

	// NewName feasibility test
	assert.EqualValues(t, customMap[NewName("psy2848048")], "World")
	_, ok := customMap[NewName("psy2848048000")]
	assert.EqualValues(t, ok, false)
}

func TestEncodingRestrictedChar(t *testing.T) {
	fmt.Println("Test5: Check restricted char in name")

	var testName Name
	// Contains '@' which is not supported
	err := testName.Init("psy gmail.com")
	fmt.Println(err)
	result := false

	if err != nil {
		fmt.Println("Error raised")
		result = true
	}
	assert.EqualValues(t, result, true)

	result = false
	stringName, err := testName.ToString()
	fmt.Println(stringName, err)
	if err != nil {
		fmt.Println("Error raised")
		result = true
	}
	assert.EqualValues(t, result, true)
}

func TestEncodingLen11ContainZero(t *testing.T) {
	fmt.Println("Test5: Check restricted char in name")

	var testName Name
	// Last letter is '0' and the length of the name is 11.
	// Last letter might be ignored if there is a bug
	testName.Init("psy28480480")

	stringName, _ := testName.ToString()
	fmt.Printf("Correct answer: 'psy28480480'\nAnswer: '%s'\n", stringName)
	assert.EqualValues(t, stringName, "psy28480480")

}

func TestNewName(t *testing.T) {
	fmt.Println("Test6: Name constructor")
	testName := NewName("bryanrhee")

	stringName, _ := testName.ToString()
	fmt.Printf("Correct answer: 'bryanrhee'\nAnswer: '%s'\n", stringName)
	assert.EqualValues(t, stringName, "bryanrhee")
}
