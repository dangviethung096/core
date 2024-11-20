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

func TryLock(ctx Context, resource string) bool {
	lockerManagerInstance.mu.Lock()
	if _, exists := lockerManagerInstance.locks[resource]; !exists {
		lockerManagerInstance.locks[resource] = &sync.Mutex{}
	}
	lockerManagerInstance.mu.Unlock()

	locked := lockerManagerInstance.locks[resource].TryLock()
	if locked {
		ctx.LogInfo("Lock resource: %s", resource)
	} else {
		ctx.LogInfo("Resource: %s is locked", resource)
	}

	return locked
}

func Lock(ctx Context, resource string) {
	lockerManagerInstance.mu.Lock()
	if _, exists := lockerManagerInstance.locks[resource]; !exists {
		lockerManagerInstance.locks[resource] = &sync.Mutex{}
	}
	lockerManagerInstance.mu.Unlock()

	lockerManagerInstance.locks[resource].Lock()
	ctx.LogInfo("Lock resource: %s", resource)
}

func Unlock(ctx Context, resource string) {
	lockerManagerInstance.mu.Lock()
	defer lockerManagerInstance.mu.Unlock()

	if lock, exists := lockerManagerInstance.locks[resource]; exists {
		lock.Unlock()
		ctx.LogInfo("Unlock resource: %s", resource)
	}
}
