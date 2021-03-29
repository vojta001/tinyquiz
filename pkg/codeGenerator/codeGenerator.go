package codeGenerator

import (
	"crypto/rand"
)

// some symbols are missing to prevent ambiguities
var alphabet = [...]byte{
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'1', '2', '3', '4', '5', '6', '7', '8', '9',
}

// Generate a code, that can be guessed after 256^randomLength attempts on average.
// randomLength can be at most 8 (= 64b of randomness)
// See GenerateCode for more details
func GenerateRandomCode(incremental uint64, randomLength uint8) ([]byte, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:randomLength]); err != nil {
		return nil, err
	}
	var random uint64
	for i := uint8(0); i < randomLength; i++ {
		random |= uint64(buf[i]) << i * 8
	}
	return GenerateCode(incremental, random), nil
}

// Generate a code (a textual identifier) that is compact and easy for humans to deal with.
// It exists for all possible values of the arguments, but the smaller they are, the shorter the code.
// The code is guaranteed to be unique, if at least one of the arguments is unique.
//
// The typical usage is to use easy to predict yet easy to make unique value for incremental
// (to ensure uniqueness) and a random value for random to make the code unpredictable.
//
// Beware the code is not guaranteed to be parsable back to incremental and random arguments.
func GenerateCode(incremental uint64, random uint64) []byte {
	var incrementalNumerals = countNumerals(incremental)
	var randomNumerals = countNumerals(random)

	var res = make([]byte, incrementalNumerals+randomNumerals)
	genCode(random, res)
	genCode(incremental, res[:incrementalNumerals])
	return res
}

// Count the bytes needed to stored encoded x
func countNumerals(x uint64) (numerals uint64) {
	for i := x; i > 0; i /= uint64(len(alphabet)) {
		numerals++
	}
	if numerals == 0 {
		numerals = 1
	}
	return
}

// Write encoded x to res from its end
func genCode(x uint64, res []byte) {
	if x == 0 {
		res[len(res)-1] = alphabet[0]
		return
	}
	for i := len(res) - 1; x > 0 && i >= 0; i-- {
		res[i] = alphabet[x%uint64(len(alphabet))]
		x /= uint64(len(alphabet))
	}
}
