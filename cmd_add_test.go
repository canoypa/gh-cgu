package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Alice", "Alice"},
		{"Alice Bob", "Alice-Bob"},
		{"alice bob", "alice-bob"},
		{"alice_bob", "alice-bob"},
		{"alice  bob", "alice-bob"},
		{"alice _ bob", "alice-bob"},
		{"-alice-", "alice"},
		{"--alice--", "alice"},
		{"日本語名前", "日本語名前"},
		{"", ""},
		{"user.name", "user-name"},
		{"a.b.c", "a-b-c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, toKey(tt.input))
		})
	}
}

func TestValidateKey(t *testing.T) {
	assert.NoError(t, validateKey("alice"))
	assert.NoError(t, validateKey("user-name"))
	assert.NoError(t, validateKey("日本語名前"))
	// ドット
	assert.ErrorContains(t, validateKey("user.name"), "invalid characters")
	assert.ErrorContains(t, validateKey("a.b.c"), "invalid characters")
	// その他の YAML 特殊文字
	assert.ErrorContains(t, validateKey("a:b"), "invalid characters")
	assert.ErrorContains(t, validateKey("a#b"), "invalid characters")
	assert.ErrorContains(t, validateKey("a b"), "invalid characters")
	// 空文字
	assert.ErrorContains(t, validateKey(""), "must not be empty")
}
