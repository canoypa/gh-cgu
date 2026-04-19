package main

// skipSyncForTesting disables background Gist sync when running tests.
// Set to true in test helpers to avoid spawning network subprocesses.
var skipSyncForTesting = false
