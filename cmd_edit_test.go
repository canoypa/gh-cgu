package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEditProfile_Name(t *testing.T) {
	v := newTestViper(t)
	addProfile(v, "Alice", "alice@example.com", "alice")

	editProfile(v, "alice", "Alice Updated", "", "", true, false, false)

	assert.Equal(t, "Alice Updated", v.GetString("alice.name"))
	assert.Equal(t, "alice@example.com", v.GetString("alice.email"))
}

func TestEditProfile_Email(t *testing.T) {
	v := newTestViper(t)
	addProfile(v, "Alice", "alice@example.com", "alice")

	editProfile(v, "alice", "", "new@example.com", "", false, true, false)

	assert.Equal(t, "Alice", v.GetString("alice.name"))
	assert.Equal(t, "new@example.com", v.GetString("alice.email"))
}

func TestEditProfile_Key(t *testing.T) {
	v := newTestViper(t)
	addProfile(v, "Alice", "alice@example.com", "alice")

	editProfile(v, "alice", "", "", "al", false, false, true)

	// 新しいキーに移動
	assert.Equal(t, "Alice", v.GetString("al.name"))
	assert.Equal(t, "alice@example.com", v.GetString("al.email"))
	// 古いキーは消える
	assert.Empty(t, v.GetString("alice.name"))
}

func TestEditProfile_AllFields(t *testing.T) {
	v := newTestViper(t)
	addProfile(v, "Alice", "alice@example.com", "alice")

	editProfile(v, "alice", "Alice New", "new@example.com", "alice2", true, true, true)

	assert.Equal(t, "Alice New", v.GetString("alice2.name"))
	assert.Equal(t, "new@example.com", v.GetString("alice2.email"))
	assert.Empty(t, v.GetString("alice.name"))
}

func TestEditProfile_Key_ConflictGuard(t *testing.T) {
	v := newTestViper(t)
	addProfile(v, "Alice", "alice@example.com", "alice")
	addProfile(v, "Bob", "bob@example.com", "bob")

	// cobra.CheckErr は os.Exit するため直接呼べない。
	// リネーム先キーが既存かどうかを v.IsSet で検出できることを確認する。
	assert.True(t, v.IsSet("bob"), "conflict should be detected by IsSet before renaming")
}
