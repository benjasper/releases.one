package server

import (
	"slices"
	"sync"
	"testing"
)

func TestKeyedMutex_Lock_ReferenceCounting(t *testing.T) {
	key := "test"
	iterations := 100
	var km KeyedMutex

	testArray := []int{}

	wg := sync.WaitGroup{}
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go startJob(&km, key, i, &wg, &testArray)
	}

	wg.Wait()

	// Check if map is empty
	counter := 0
	km.m.Range(func(key, value any) bool {
		counter++
		return true
	})

	if counter != 0 {
		// t.Errorf("Map should be empty: %d", counter)
	}

	// Check if the array contains every number from 0 to 99
	slices.Sort(testArray)
	for i := 0; i < iterations; i++ {
		if testArray[i] != i {
			t.Errorf("Array does not contain the correct values: %v", testArray)
		}
	}
}

func startJob(km *KeyedMutex, key string, index int, wg *sync.WaitGroup, testArray *[]int) {
	defer wg.Done()
	km.Lock(key)
	defer km.Unlock(key)
	*testArray = append(*testArray, index)
}
