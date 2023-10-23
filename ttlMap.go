/*
ttlMap.go
-John Taylor
2023-10-21

ttlMap is a "time-to-live" map such that after a given amount of time, items
in the map are deleted.

When a Put() occurs, the lastAccess time is set to time.Now().Unix()
When a Get() occurs, the lastAccess time is updated to time.Now().Unix()
Therefore, only items that are not called by Get() will be deleted after the TTL occurs.

Adopted from: https://stackoverflow.com/a/25487392/452281

Changes from the referenced implementation
==========================================
1) the may key is user definable by setting CustomKeyType (defaults to string)
2) use interface{} instead of string as the map value so that any data type can be used
3) added All() function
4) use item.Value instead of item.value so that it can be externally referenced
5) added user configurable prune interval - search for expired items every 'pruneInterval' seconds
6) toggle for refreshLastAccessOnGet - update item's lastAccessTime on a .Get() when set to true

*/

package ttlMap

import (
	"sync"
	"time"
)

const version string = "1.1.0"

type CustomKeyType string

type item struct {
	Value      interface{}
	lastAccess int64
}

type ttlMap struct {
	m       map[CustomKeyType]*item
	l       sync.Mutex
	refresh bool
}

func New(maxTTL int, ln int, pruneInterval int, refreshLastAccessOnGet bool) (m *ttlMap) {
	// if pruneInterval > maxTTL {
	// 	print("WARNING: ttlMap: pruneInterval > maxTTL\n")
	// }
	m = &ttlMap{m: make(map[CustomKeyType]*item, ln)}
	m.refresh = refreshLastAccessOnGet
	go func() {
		for now := range time.Tick(time.Second * time.Duration(pruneInterval)) {
			currentTime := now.Unix()
			m.l.Lock()
			for k, v := range m.m {
				// print("TICK:", currentTime, "  ", v.lastAccess, "  ", (currentTime - v.lastAccess), "  ", maxTTL, "  ", k, "\n")
				if currentTime-v.lastAccess >= int64(maxTTL) {
					delete(m.m, k)
					// print("deleting: ", k, "\n")
				}
			}
			// print("\n")
			m.l.Unlock()
		}
	}()
	return
}

func (m *ttlMap) Len() int {
	return len(m.m)
}

func (m *ttlMap) Put(k CustomKeyType, v interface{}) {
	m.l.Lock()
	it, ok := m.m[k]
	if !ok {
		it = &item{Value: v}
		m.m[k] = it
	}
	it.lastAccess = time.Now().Unix()
	m.l.Unlock()
}

func (m *ttlMap) Get(k CustomKeyType) (v interface{}) {
	m.l.Lock()
	if it, ok := m.m[k]; ok {
		v = it.Value
		if m.refresh {
			m.m[k].lastAccess = time.Now().Unix()
		}
	}
	m.l.Unlock()
	return
}

func (m *ttlMap) Delete(k CustomKeyType) bool {
	m.l.Lock()
	_, ok := m.m[k]
	if !ok {
		m.l.Unlock()
		return false
	}
	delete(m.m, k)
	m.l.Unlock()
	return true
}

func (m *ttlMap) All() map[CustomKeyType]*item {
	return m.m
}
