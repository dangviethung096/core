package core

import (
	"fmt"
	"hash/fnv"
)

type PgLock struct {
	session dbSession
	lockID  int64
	lockKey string
}

func NewPgLock(session dbSession, lockKey string) *PgLock {
	lockID := GetLockIdFromLockKey(lockKey)
	return &PgLock{
		session: session,
		lockKey: lockKey,
		lockID:  lockID,
	}
}

func (m *PgLock) Lock() Error {
	var success bool
	err := m.session.QueryRow("SELECT pg_try_advisory_lock($1)", m.lockID).Scan(&success)
	if err != nil {
		return NewError(ERROR_CODE_FROM_DATABASE, fmt.Sprintf("failed to acquire lock: %v", err))
	}
	if !success {
		return NewError(ERROR_CODE_FROM_DATABASE, "lock is already held")
	}
	return nil
}

func (m *PgLock) Unlock() Error {
	var success bool
	err := m.session.QueryRow("SELECT pg_advisory_unlock($1)", m.lockID).Scan(&success)
	if err != nil {
		return NewError(ERROR_CODE_FROM_DATABASE, fmt.Sprintf("failed to release lock: %v", err))
	}
	if !success {
		return NewError(ERROR_CODE_FROM_DATABASE, "lock was not held")
	}
	return nil
}

// HashStringToInt hashes a string to an int using FNV-1a algorithm
func GetLockIdFromLockKey(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}
