package core

import (
	"sync"
)

type lockManager struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func newLockManager() *lockManager {
	return &lockManager{
		locks: make(map[string]*sync.Mutex),
	}
}

func TryLock(resource string) bool {
	lockerManagerInstance.mu.Lock()
	if _, exists := lockerManagerInstance.locks[resource]; !exists {
		lockerManagerInstance.locks[resource] = &sync.Mutex{}
	}
	lockerManagerInstance.mu.Unlock()

	return lockerManagerInstance.locks[resource].TryLock()
}

func Lock(resource string) {
	lockerManagerInstance.mu.Lock()
	if _, exists := lockerManagerInstance.locks[resource]; !exists {
		lockerManagerInstance.locks[resource] = &sync.Mutex{}
	}
	lockerManagerInstance.mu.Unlock()

	lockerManagerInstance.locks[resource].Lock()
}

func Unlock(resource string) {
	lockerManagerInstance.mu.Lock()
	defer lockerManagerInstance.mu.Unlock()

	if lock, exists := lockerManagerInstance.locks[resource]; exists {
		lock.Unlock()
	}
}
