package common

import "sync"

type retryLimiter struct {
	record   map[string]int
	maxRetry int
	mux      sync.RWMutex
}

type RetryLimiter interface {
	Add(id string)
	Try(id string) bool
	GetMaxRetryTimes() int
	GetCurrentRetryTimes(id string) int
}

func NewRetryLimiter(maxRetry int) RetryLimiter {
	return &retryLimiter{
		record:   map[string]int{},
		maxRetry: maxRetry,
		mux:      sync.RWMutex{},
	}
}

func (r *retryLimiter) Add(id string) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.record[id]++
}

func (r *retryLimiter) Try(id string) bool {
	r.mux.RLock()
	defer r.mux.RUnlock()
	return r.maxRetry == 0 || r.record[id] <= r.maxRetry
}

func (r *retryLimiter) GetMaxRetryTimes() int {
	return r.maxRetry
}

func (r *retryLimiter) GetCurrentRetryTimes(id string) int {
	return r.record[id]
}
