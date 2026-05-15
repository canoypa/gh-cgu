//go:build windows

package main

// acquireSyncLock on Windows always succeeds (no-op).
// LockFileEx-based advisory locking is not implemented.
// Background sync is disabled on Windows in syncToGist() to avoid
// concurrent config corruption until proper locking is added.
func acquireSyncLock() (release func(), ok bool) {
	return func() {}, true
}
