package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPascal(t *testing.T) {
	cases := [][]string{
		// basic
		[]string{"test", "Test"},

		// common separators
		[]string{"test-case", "TestCase"},
		[]string{"test_case", "TestCase"},
		[]string{"test case", "TestCase"},

		// numbers
		[]string{"test1case", "Test1Case"},
		[]string{"1test", "1Test"},

		// mixed case
		[]string{"tEsTiNg", "TEsTiNg"},

		// special characters
		[]string{"хлеб", "Хлеб"},
	}

	for _, c := range cases {
		assert.Equal(t, ToPascalCase(c[0]), c[1])
	}
}
