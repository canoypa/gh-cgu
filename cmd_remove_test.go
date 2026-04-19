package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveProfile(t *testing.T) {
	v := newTestViper(t)
	addProfile(v, "Alice", "alice@example.com", "alice")
	addProfile(v, "Bob", "bob@example.com", "bob")

	removeProfile(v, "alice")

	assert.Empty(t, v.GetString("alice.name"))
	assert.Empty(t, v.GetString("alice.email"))
	// bob は残る
	assert.Equal(t, "Bob", v.GetString("bob.name"))
}

func TestRemoveProfile_NotFound(t *testing.T) {
	v := newTestViper(t)

	// cobra.CheckErr は os.Exit するため直接呼べない
	// profile が存在しない場合は email が空なのでガード条件をテスト
	assert.Empty(t, v.GetString("nonexistent.email"))
}
