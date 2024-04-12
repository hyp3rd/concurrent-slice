package cslice_test

import (
	"sync"
	"testing"

	cslice "https://github.com/hyp3rd/concurrent-slice/pkg"
)

// TestSetAndGet tests the basic Set and Get methods for correctness.
func TestSetAndGet(t *testing.T) {
	cs := cslice.NewConcurrentSlice()
	cs.Set("hello")
	cs.Set("world")

	if item, ok := cs.Get(0); !ok || item != "hello" {
		t.Errorf("Expected 'hello', got '%v'", item)
	}
	if item, ok := cs.Get(1); !ok || item != "world" {
		t.Errorf("Expected 'world', got '%v'", item)
	}
}

// TestConcurrency tests the ConcurrentSlice for safe concurrent access.
func TestConcurrency(t *testing.T) {
	cs := cslice.NewConcurrentSlice()
	var wg sync.WaitGroup

	// Start several goroutines that add items to the slice concurrently.
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			cs.Set(val)
		}(i)
	}

	wg.Wait()

	if count := cs.Count(); count != 100 {
		t.Errorf("Expected 100 items, got %d", count)
	}
}

// TestDelete tests the deletion functionality.
func TestDelete(t *testing.T) {
	cs := cslice.NewConcurrentSlice()
	items := []interface{}{"a", "b", "c", "d", "e"}
	cs.SetMany(items...)

	cs.Delete(2) // Remove "c"

	expected := []interface{}{"a", "b", "d", "e"}
	for i, exp := range expected {
		if item, ok := cs.Get(i); !ok || item != exp {
			t.Errorf("After Delete, expected %v at index %d, got %v", exp, i, item)
		}
	}
}

// TestIter tests the iteration functionality.
func TestIter(t *testing.T) {
	cs := cslice.NewConcurrentSlice()
	items := []interface{}{"one", "two", "three"}
	cs.SetMany(items...)

	ch := cs.Iter(false)
	index := 0
	for item := range ch {
		if items[index] != item.Value {
			t.Errorf("Iter error: expected %v, got %v", items[index], item.Value)
		}
		index++
	}
}

// TestClear tests the clear functionality.
func TestClear(t *testing.T) {
	cs := cslice.NewConcurrentSlice()
	cs.SetMany("a", "b", "c")
	cs.Clear()

	if count := cs.Count(); count != 0 {
		t.Errorf("Expected 0 items after Clear, got %d", count)
	}
}

// TestContains checks if the Contains method works as expected.
func TestContains(t *testing.T) {
	cs := cslice.NewConcurrentSlice()
	cs.Set("hello")
	cs.Set("world")

	if !cs.Contains("world") {
		t.Errorf("Contains failed to find 'world'")
	}
	if cs.Contains("missing") {
		t.Errorf("Contains found 'missing' which was not expected")
	}
}
