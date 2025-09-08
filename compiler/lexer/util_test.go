package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtils(t *testing.T) {
	t.Run("isLetter", func(t *testing.T) {
		t.Run("Should return true for letters", func(t *testing.T) {
			assert.True(t, isLetter('a'))
			assert.True(t, isLetter('z'))
			assert.True(t, isLetter('A'))
			assert.True(t, isLetter('Z'))
			assert.True(t, isLetter('_'))
		})

		t.Run("Should return false for non-letters", func(t *testing.T) {
			assert.False(t, isLetter('1'))
			assert.False(t, isLetter(' '))
			assert.False(t, isLetter('.'))
			assert.False(t, isLetter('!'))
			assert.False(t, isLetter('@'))
		})
	})

	t.Run("isDigit", func(t *testing.T) {
		t.Run("Should return true for digits", func(t *testing.T) {
			assert.True(t, isDigit('0'))
			assert.True(t, isDigit('9'))
		})

		t.Run("Should return false for non-digits", func(t *testing.T) {
			assert.False(t, isDigit('a'))
			assert.False(t, isDigit(' '))
			assert.False(t, isDigit('.'))
			assert.False(t, isDigit('!'))
			assert.False(t, isDigit('@'))
		})
	})

	t.Run("isWhitespace", func(t *testing.T) {
		t.Run("Should return true for whitespace", func(t *testing.T) {
			assert.True(t, isWhitespace(' '))
			assert.True(t, isWhitespace('\t'))
			assert.True(t, isWhitespace('\r'))
		})

		t.Run("Should return false for non-whitespace", func(t *testing.T) {
			assert.False(t, isWhitespace('a'))
			assert.False(t, isWhitespace('.'))
			assert.False(t, isWhitespace('!'))
			assert.False(t, isWhitespace('@'))
		})
	})

	t.Run("containsDot", func(t *testing.T) {
		t.Run("Should return true for strings containing dots", func(t *testing.T) {
			assert.True(t, containsDot("1.2.3"))
			assert.True(t, containsDot("1.2.3.4.5.6.7.8.9.0"))
		})

		t.Run("Should return false for strings not containing dots", func(t *testing.T) {
			assert.False(t, containsDot("123"))
			assert.False(t, containsDot("1234567890"))
		})
	})
}
