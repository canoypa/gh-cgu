package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestViper creates a viper instance backed by a temporary yaml file.
func newTestViper(t *testing.T) *viper.Viper {
	t.Helper()
	skipSyncForTesting = true
	t.Cleanup(func() { skipSyncForTesting = false })
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "test-config.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte{}, 0600))

	v := viper.New()
	v.SetConfigFile(cfgFile)
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadInConfig())
	return v
}

func TestAddProfile(t *testing.T) {
	v := newTestViper(t)

	addProfile(v, "Alice", "alice@example.com", "alice")

	assert.Equal(t, "Alice", v.GetString("alice.name"))
	assert.Equal(t, "alice@example.com", v.GetString("alice.email"))
}

func TestAddProfile_MultipleProfiles(t *testing.T) {
	v := newTestViper(t)

	addProfile(v, "Alice", "alice@example.com", "alice")
	addProfile(v, "Bob", "bob@example.com", "bob")

	assert.Equal(t, "Alice", v.GetString("alice.name"))
	assert.Equal(t, "Bob", v.GetString("bob.name"))
}
