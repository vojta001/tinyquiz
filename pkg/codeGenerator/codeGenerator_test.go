package codeGenerator

import (
	"bytes"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	test := func(incremental uint64, random uint64, expected []byte) {
		if actual := GenerateCode(incremental, random); !bytes.Equal(actual, expected) {
			t.Errorf("GenerateCode(%d, %d) returned %#v while %#v was expected", incremental, random, actual, expected)
		} else if len(actual) != cap(actual) {
			t.Errorf("GenerateCode(%d, %d) returned slice with capacity %d, while its length is %d. Potential memory waste", incremental, random, cap(actual), len(actual))
		}
	}
	test(0, 0, []byte{'A', 'A'})
	test(32, 31, []byte{'B', 'A', '9'})
	const maxUint64 = ^uint64(0)
	test(maxUint64, maxUint64, []byte{'S', '9', '9', '9', '9', '9', '9', '9', '9', '9', '9', '9', '9', 'S', '9', '9', '9', '9', '9', '9', '9', '9', '9', '9', '9', '9'})
}
