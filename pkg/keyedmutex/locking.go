package keyedmutex

import (
	"fmt"
	"sync"
)

type refCountedMutex struct {
	mtx   sync.Mutex
	count int
}

type KeyedMutex struct {
	mu sync.Mutex
	m  map[string]*refCountedMutex
}

// NewKeyedMutex initializes a new KeyedMutex.
func NewKeyedMutex() *KeyedMutex {
	return &KeyedMutex{
		m: make(map[string]*refCountedMutex),
	}
}

// Lock acquires the mutex for a given key.
func (km *KeyedMutex) Lock(key string) {
	km.mu.Lock()
	rmtx, exists := km.m[key]
	if !exists {
		rmtx = &refCountedMutex{}
		km.m[key] = rmtx
	}
	rmtx.count++
	km.mu.Unlock()

	// Lock the underlying mutex
	rmtx.mtx.Lock()
}

// Unlock releases the mutex for a given key.
func (km *KeyedMutex) Unlock(key string) {
	km.mu.Lock()
	rmtx, exists := km.m[key]
	if !exists {
		km.mu.Unlock()
		panic(fmt.Sprintf("attempt to unlock a non-existent key: %s", key))
	}

	rmtx.count--
	if rmtx.count == 0 {
		// Delete the key if no more references
		delete(km.m, key)
	}
	km.mu.Unlock()

	// Unlock the underlying mutex
	rmtx.mtx.Unlock()
}
