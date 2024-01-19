package common

import (
	"sync"
)

const (
	OperationPendingFmt = "already an operation for the specified resource %s"
)

type ResourceLocks struct {
	locks map[string]struct{}
	mux   sync.Mutex
}

func NewResourceLocks() *ResourceLocks {
	return &ResourceLocks{
		locks: make(map[string]struct{}),
	}
}

func (lock *ResourceLocks) TryAcquire(resourceID string) bool {
	lock.mux.Lock()
	defer lock.mux.Unlock()
	_, acquired := lock.locks[resourceID]
	if acquired {
		return false
	}
	lock.locks[resourceID] = struct{}{}
	return true
}

func (lock *ResourceLocks) Release(resourceID string) {
	lock.mux.Lock()
	defer lock.mux.Unlock()
	delete(lock.locks, resourceID)
}
