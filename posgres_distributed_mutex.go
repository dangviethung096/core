package core

import (
	"fmt"
)

type PgMutex struct {
	session dbSession
	lockID  int64
}

func NewPgMutex(session dbSession, lockID int64) *PgMutex {
	return &PgMutex{
		session: session,
		lockID:  lockID,
	}
}

func (m *PgMutex) Reserve() Error {
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

func (m *PgMutex) Release() Error {
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
