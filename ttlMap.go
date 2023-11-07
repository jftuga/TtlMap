/*
TtlMap.go
-John Taylor
2023-10-21

TtlMap is a "time-to-live" map such that after a given amount of time, items
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
7) add Delete() and Clear() functions

*/

package TtlMap

import (
	"maps"
	"sync"
	"time"
)

const version string = "1.3.0"

type CustomKeyType string

type item struct {
	Value      interface{}
	lastAccess int64
}

type TtlMap struct {
	m       map[CustomKeyType]*item
	l       sync.Mutex
	refresh bool
	stop    chan bool
}

func New(maxTTL int, ln int, pruneInterval int, refreshLastAccessOnGet bool) (m *TtlMap) {
	// if pruneInterval > maxTTL {
	// 	print("WARNING: TtlMap: pruneInterval > maxTTL\n")
	// }
	m = &TtlMap{m: make(map[CustomKeyType]*item, ln), stop: make(chan bool)}
	m.refresh = refreshLastAccessOnGet
	go func() {
		for {
			select {
			case <-m.stop:
				return
			case now := <-time.Tick(time.Second * time.Duration(pruneInterval)):
				currentTime := now.Unix()
				m.l.Lock()
				for k, v := range m.m {
					//print("TICK:", currentTime, "  ", v.lastAccess, "  ", (currentTime - v.lastAccess), "  ", maxTTL, "  ", k, "\n")
					if currentTime-v.lastAccess >= int64(maxTTL) {
						delete(m.m, k)
						// print("deleting: ", k, "\n")
					}
				}
				// print("\n")
				m.l.Unlock()
			}
		}
	}()
	return
}

func (m *TtlMap) Len() int {
	return len(m.m)
}

func (m *TtlMap) Put(k CustomKeyType, v interface{}) {
	m.l.Lock()
	it, ok := m.m[k]
	if !ok {
		it = &item{Value: v}
		m.m[k] = it
	}
	it.lastAccess = time.Now().Unix()
	m.l.Unlock()
}

func (m *TtlMap) Get(k CustomKeyType) (v interface{}) {
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

func (m *TtlMap) Delete(k CustomKeyType) bool {
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

func (m *TtlMap) Clear() {
	m.l.Lock()
	clear(m.m)
	m.l.Unlock()
}

func (m *TtlMap) All() map[CustomKeyType]*item {
	m.l.Lock()
	dst := make(map[CustomKeyType]*item, len(m.m))
	maps.Copy(dst, m.m)
	m.l.Unlock()
	return dst
}

func (m *TtlMap) Close() {
	m.stop <- true
}
