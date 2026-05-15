//go:build !windows

package main

import (
	"os"
	"syscall"
)

// acquireSyncLock tries to acquire a non-blocking exclusive advisory lock on
// the sync lock file. Returns (release, true) on success. If another process
// already holds the lock, returns (nil, false) and the caller should abort.
func acquireSyncLock() (release func(), ok bool) {
	path := syncLockPath()
	if path == "" {
		return func() {}, true
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		// Cannot create lock file; proceed without lock.
		return func() {}, true
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, false // Another sync is already running
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, true
}
