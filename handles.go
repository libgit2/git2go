package git

import (
	"fmt"
	"sync"
	"unsafe"
)

type HandleList struct {
	sync.RWMutex
	// stores the Go pointers
	handles []interface{}
	// indicates which indices are in use
	set map[uintptr]bool
}

func NewHandleList() *HandleList {
	return &HandleList{
		handles: make([]interface{}, 5),
		set:     make(map[uintptr]bool),
	}
}

// findUnusedSlot finds the smallest-index empty space in our
// list. You must only run this function while holding a write lock.
func (v *HandleList) findUnusedSlot() uintptr {
	for i := 1; i < len(v.handles); i++ {
		isUsed := v.set[uintptr(i)]
		if !isUsed {
			return uintptr(i)
		}
	}

	// reaching here means we've run out of entries so append and
	// return the new index, which is equal to the old length.
	slot := len(v.handles)
	v.handles = append(v.handles, nil)

	return uintptr(slot)
}

// Track adds the given pointer to the list of pointers to track and
// returns a pointer value which can be passed to C as an opaque
// pointer.
func (v *HandleList) Track(pointer interface{}) unsafe.Pointer {
	v.Lock()

	slot := v.findUnusedSlot()
	v.handles[slot] = pointer
	v.set[slot] = true

	v.Unlock()

	return unsafe.Pointer(slot)
}

// Untrack stops tracking the pointer given by the handle
func (v *HandleList) Untrack(handle unsafe.Pointer) {
	slot := uintptr(handle)

	v.Lock()

	v.handles[slot] = nil
	delete(v.set, slot)

	v.Unlock()
}

// Get retrieves the pointer from the given handle
func (v *HandleList) Get(handle unsafe.Pointer) interface{} {
	slot := uintptr(handle)

	v.RLock()

	if _, ok := v.set[slot]; !ok {
		panic(fmt.Sprintf("invalid pointer handle: %p", handle))
	}

	ptr := v.handles[slot]

	v.RUnlock()

	return ptr
}
