package state

import "sync"

// Base is the base of all events.
type Base struct {
	vars    map[string]interface{}
	varsMut *sync.RWMutex
}

func NewBase() *Base {
	return &Base{
		vars:    make(map[string]interface{}),
		varsMut: new(sync.RWMutex),
	}
}

// Set stores the passed element under the given key.
func (b *Base) Set(key string, val interface{}) {
	b.varsMut.Lock()
	b.vars[key] = val
	b.varsMut.Unlock()
}

// Get gets the element with the passed key.
func (b *Base) Get(key string) (val interface{}) {
	val, _ = b.Lookup(key)
	return val
}

// Lookup returns the element with the passed key.
// Additionally, it specifies with the second return parameter, if the element
// exists, acting similar to a two parameter map lookup.
func (b *Base) Lookup(key string) (val interface{}, ok bool) {
	b.varsMut.RLock()
	defer b.varsMut.RUnlock()

	val, ok = b.vars[key]
	return
}
