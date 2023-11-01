package common

import (
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

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

// GenerateRandIdSuffix generates a random resource id.
func GenerateHashInEightBytes(input string) string {
	h := fnv.New32a()
	h.Write([]byte(input))
	return fmt.Sprintf("%.8x", h.Sum32())
}

func EntryFunction(functionName string) (info string, hash string) {
	current := time.Now()
	hash = GenerateHashInEightBytes(current.UTC().String())
	return fmt.Sprintf("*************** enter %s at %s hash %s ***************", functionName,
		current.Format(DefaultTimeFormat), hash), hash
}

// ExitFunction print timestamps
func ExitFunction(functionName, hash string) (info string) {
	current := time.Now()
	return fmt.Sprintf("=============== exit %s at %s hash %s ===============", functionName,
		current.Format(DefaultTimeFormat), hash)
}
