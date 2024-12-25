package server

import (
	"fmt"
	"sync"
)

type refCountedMutex struct {
	mtx   *sync.Mutex
}

type KeyedMutex struct {
	m sync.Map
}

func (km *KeyedMutex) Lock(key string) {
	val, _ := km.m.LoadOrStore(key, &refCountedMutex{
		mtx:   &sync.Mutex{},
	})
	rmtx := val.(*refCountedMutex)

	rmtx.mtx.Lock()
}

func (km *KeyedMutex) Unlock(key string) {
	val, ok := km.m.Load(key)
	if !ok {
		panic(fmt.Sprintf("attempt to unlock a non-existent key: %s", key))
	}
	rmtx := val.(*refCountedMutex)

	rmtx.mtx.Unlock()
}

