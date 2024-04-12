package cslice

import (
	"fmt"
	"sync"
)

// ConcurrentSlice type that can be safely shared between goroutines
type ConcurrentSlice struct {
	mutex sync.RWMutex
	items []interface{}
}

// ConcurrentSliceItem is the type of a concurrent slice item
type ConcurrentSliceItem struct {
	Index int
	Value interface{}
}

// NewConcurrentSlice Creates a new concurrent slice
func NewConcurrentSlice() *ConcurrentSlice {
	return &ConcurrentSlice{
		items: make([]interface{}, 0),
	}
}

// Set Appends an item to the concurrent slice
func (cs *ConcurrentSlice) Set(item interface{}) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.items = append(cs.items, item)
}

// SetAt sets an item at a specific index in the concurrent slice and returns an error if out of bounds.
func (cs *ConcurrentSlice) SetAt(index int, item interface{}) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if index < 0 || index >= len(cs.items) {
		return fmt.Errorf("index %d out of range", index)
	}
	cs.items[index] = item
	return nil
}

// SetMany appends multiple items to the concurrent slice
func (cs *ConcurrentSlice) SetMany(items ...interface{}) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.items = append(cs.items, items...)
}

// Get Gets an item from the concurrent slice
func (cs *ConcurrentSlice) Get(index int) (interface{}, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	if index < 0 || index >= len(cs.items) {
		return nil, false
	}
	return cs.items[index], true
}

// Delete removes an item from the concurrent slice efficiently.
func (cs *ConcurrentSlice) Delete(index int) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if index < 0 || index >= len(cs.items) {
		return
	}
	cs.items = append(cs.items[:index], cs.items[index+1:]...)
	cs.items = append([]interface{}(nil), cs.items...) // Truncate slice to free memory if necessary
}

// Count returns the number of items in the slice
func (cs *ConcurrentSlice) Count() int {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return len(cs.items)
}

// Clear removes all items from the concurrent slice.
func (cs *ConcurrentSlice) Clear() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.items = nil // or cs.items = make([]interface{}, 0) to reset the slice without freeing the underlying array
}

// Contains checks if the slice contains an item.
func (cs *ConcurrentSlice) Contains(item interface{}) bool {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	for _, v := range cs.items {
		if v == item {
			return true
		}
	}
	return false
}

// Iter iterates over the items in the concurrent slice. It can optionally use a buffered channel.
func (cs *ConcurrentSlice) Iter(buffered bool) <-chan ConcurrentSliceItem {
	var c chan ConcurrentSliceItem
	if buffered {
		cs.mutex.RLock()
		c = make(chan ConcurrentSliceItem, len(cs.items))
		cs.mutex.RUnlock()
	} else {
		c = make(chan ConcurrentSliceItem)
	}

	go func() {
		cs.mutex.RLock()
		defer cs.mutex.RUnlock()
		defer close(c)

		for index, value := range cs.items {
			c <- ConcurrentSliceItem{index, value}
		}
	}()

	return c
}

// IterFunc type for handling items during iteration.
type IterFunc func(item ConcurrentSliceItem) bool

// IterWithFunc iterates over the items in the concurrent slice and allows early termination.
func (cs *ConcurrentSlice) IterWithFunc(buffered bool, fn IterFunc) {
	var c chan ConcurrentSliceItem
	if buffered {
		cs.mutex.RLock()
		c = make(chan ConcurrentSliceItem, len(cs.items))
		cs.mutex.RUnlock()
	} else {
		c = make(chan ConcurrentSliceItem)
	}

	go func() {
		cs.mutex.RLock()
		defer cs.mutex.RUnlock()
		defer close(c)

		for index, value := range cs.items {
			if !fn(ConcurrentSliceItem{index, value}) {
				break // Stop iteration based on the function's return value
			}
			c <- ConcurrentSliceItem{index, value}
		}
	}()
}
