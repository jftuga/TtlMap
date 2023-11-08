/*
TtlMap.go
-John Taylor
2023-10-21

TtlMap is a "time-to-live" map such that after a given amount of time, items
in the map are deleted.

When a Put() occurs, the lastAccess time is set to time.Now().Unix()
When a Get() occurs, the lastAccess time is updated to time.Now().Unix()
Therefore, only items that are not called by Get() will be deleted after the TTL occurs.
A GetNoUpdate() can be used in which case the lastAccess time will NOT be updated.

Adopted from: https://stackoverflow.com/a/25487392/452281
*/

package TtlMap

import (
	"maps"
	"sync"
	"time"
)

const version string = "1.5.0"

type CustomKeyType interface {
	comparable
}

type item struct {
	Value      interface{}
	lastAccess int64
}

type TtlMap[T CustomKeyType] struct {
	m       map[T]*item
	l       sync.Mutex
	refresh bool
	stop    chan bool
}

func New[T CustomKeyType](maxTTL time.Duration, ln int, pruneInterval time.Duration, refreshLastAccessOnGet bool) (m *TtlMap[T]) {
	// if pruneInterval > maxTTL {
	// 	print("WARNING: TtlMap: pruneInterval > maxTTL\n")
	// }
	m = &TtlMap[T]{m: make(map[T]*item, ln), stop: make(chan bool)}
	m.refresh = refreshLastAccessOnGet
	maxTTL /= 1000000000
	// print("maxTTL: ", maxTTL, "\n")
	go func() {
		for {
			select {
			case <-m.stop:
				return
			case now := <-time.Tick(pruneInterval):
				currentTime := now.Unix()
				m.l.Lock()
				for k, v := range m.m {
					// print("TICK:", currentTime, "  ", v.lastAccess, "  ", currentTime-v.lastAccess, "  ", maxTTL, "  ", k, "\n")
					if currentTime-v.lastAccess >= int64(maxTTL) {
						delete(m.m, k)
						// print("deleting: ", k, "\n")
					}
				}
				// print("----\n")
				m.l.Unlock()
			}
		}
	}()
	return
}

func (m *TtlMap[T]) Len() int {
	return len(m.m)
}

func (m *TtlMap[T]) Put(k T, v interface{}) {
	m.l.Lock()
	it, ok := m.m[k]
	if !ok {
		it = &item{Value: v}
		m.m[k] = it
	}
	it.lastAccess = time.Now().Unix()
	m.l.Unlock()
}

func (m *TtlMap[T]) Get(k T) (v interface{}) {
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

func (m *TtlMap[T]) GetNoUpdate(k T) (v interface{}) {
	m.l.Lock()
	if it, ok := m.m[k]; ok {
		v = it.Value
	}
	m.l.Unlock()
	return
}

func (m *TtlMap[T]) Delete(k T) bool {
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

func (m *TtlMap[T]) Clear() {
	m.l.Lock()
	clear(m.m)
	m.l.Unlock()
}

func (m *TtlMap[T]) All() map[T]*item {
	m.l.Lock()
	dst := make(map[T]*item, len(m.m))
	maps.Copy(dst, m.m)
	m.l.Unlock()
	return dst
}

func (m *TtlMap[T]) Close() {
	m.stop <- true
}
