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
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, toKey(tt.input))
		})
	}
}
