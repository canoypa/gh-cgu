package main

// skipSyncForTesting disables background Gist sync when running tests.
// Set to true in test helpers to avoid spawning network subprocesses.
var skipSyncForTesting = false

// metaFileOverride redirects meta file reads/writes to a custom path in tests.
// Empty string means use the default path.
var metaFileOverride = ""
