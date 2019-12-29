package types

import (
	"bytes"
	"errors"
	"strings"
)

// Might be relocated into 'types'

const charmap string = "0123456789abcdefghijhlmnopqrstuvwxyz-._"
const encodingBase int = len(charmap)

// Name is uint128 datatype, and supports redable ID which contains up to 20 letters.
// But, Golang has no uint128 in native, so it should be created manually.
// ID fills H uint64 first 10 letters, and fills L next 10 letters.
type Name struct {
	H, L uint64
}

// charToCode encodes unit character by Base39 encoding rules
// Char '0' matches uint64 1, and other mathces accordingly
func charToCode(letter rune) (uint64, error) {
	if !strings.Contains(charmap, string(letter)) {
		return ^uint64(0), errors.New("The char of a name can be used within alphanumeric and lower case character")
	}

	if '0' <= letter && letter <= '9' {
		return uint64(letter) - '0' + 1, nil
	} else if 'a' <= letter && letter <= 'z' {
		return uint64(letter) - 'a' + 10 + 1, nil
	} else {
		// Can return only by this method,
		// but afraid that strings.Index() consume some exec time.
		return uint64(strings.Index(charmap[36:], string(letter)) + 36 + 1), nil
	}
	//return uint64( strings.Index(charmap, string(letter) )), nil
}

// Encoding for uint64
// Attatch it to Name struct
func partialStringNameToCode(partialStringName string) uint64 {
	if len(partialStringName) > 10 {
		return ^uint64(0)
	}

	var result uint64
	for _, letter := range partialStringName {
		letterCode, err := charToCode(letter)
		if err == nil {
			result = result<<6 + letterCode
		} else {
			return ^uint64(0)
		}
	}
	return result
}

// uint64 Code -> String
func partialCodeToString(partialName uint64) (string, error) {
	// Error == ^uint64(0)
	if partialName == ^uint64(0) {
		return "", errors.New("Invalid name")
	}

	// Zero value -> No string contained
	if partialName == 0 {
		return "", nil
	}

	// The fastest string concat method:
	// https://stackoverflow.com/questions/1760757/how-to-efficiently-concatenate-strings-in-go
	result := make([]byte, 10)
	intermediateDiv := partialName
	cnt := 0
	for {
		cnt++

		intermediateDiv2, index := intermediateDiv>>6, intermediateDiv%64-1
		result[10-cnt] = charmap[index]
		intermediateDiv = intermediateDiv2

		if intermediateDiv == 0 || cnt > 10 {
			break
		}
	}

	// Remove useless trailing \x00
	result = bytes.Trim(result, "\x00")
	return string(result), nil
}

// NewName acts like a constuctor of "Name"
func NewName(stringName string) Name {
	var nameObj Name
	nameObj.Init(stringName)

	return nameObj
}

// Init work as a initializer of Name
func (N *Name) Init(stringName string) error {
	if len(stringName) > 20 {
		return errors.New("Name cannot exceed more than 20 characters")
	}
	lowerCasedStringName := strings.ToLower(stringName)
	if len(stringName) <= 10 {
		result := partialStringNameToCode(lowerCasedStringName)
		if result == ^uint64(0) {
			return errors.New("Name can only contain 0-9 a-z .-_")
		}
		N.H = result
	} else {
		partialStringNameHigh := lowerCasedStringName[:10]
		partialStringNameLow := lowerCasedStringName[10:]

		result1 := partialStringNameToCode(partialStringNameHigh)
		result2 := partialStringNameToCode(partialStringNameLow)
		if result1 == ^uint64(0) || result2 == ^uint64(0) {
			return errors.New("Name can only contain 0-9 a-z .-_")
		}
		N.H = result1
		N.L = result2
	}
	return nil
}

// ToString extracts string form from uint128 "Name" struct
func (N *Name) ToString() (string, error) {
	upperString, err1 := partialCodeToString(N.H)
	lowerString, err2 := partialCodeToString(N.L)
	if err1 == nil && err2 == nil && upperString != "" {
		return upperString + lowerString, nil
	}
	return "", errors.New("Name object is not initialized")
}
