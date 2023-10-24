package common

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

type ResourceLocks struct {
	locks sets.String
	mux   sync.Mutex
}

func NewResourceLocks() *ResourceLocks {
	return &ResourceLocks{
		locks: sets.NewString(),
	}
}
