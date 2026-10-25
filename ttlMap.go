/*
TtlMap.go
-John Taylor
2023-10-21

TtlMap is a "time-to-live" map such that after a given amount of time, items
in the map are deleted.

When a Put() occurs, the lastAccess time is set to time.Now()
When a Get() occurs with refresh enabled, the lastAccess time is updated to time.Now()
Therefore, only items that are not called by Get() will be deleted after the TTL occurs.
A GetNoUpdate() can be used in which case the lastAccess time will NOT be updated.

Adopted from: https://stackoverflow.com/a/25487392/452281
*/

package TtlMap

import (
	"sync"
	"time"
)

const version string = "1.6.0"

type CustomKeyType interface {
	comparable
}

type item struct {
	Value      interface{}
	lastAccess time.Time
}

type TtlMap[T CustomKeyType] struct {
	m         map[T]*item
	l         sync.Mutex
	refresh   bool
	done      chan struct{}
	closeOnce sync.Once
}

func New[T CustomKeyType](maxTTL time.Duration, ln int, pruneInterval time.Duration, refreshLastAccessOnGet bool) (m *TtlMap[T]) {
	m = &TtlMap[T]{m: make(map[T]*item, ln), done: make(chan struct{})}
	m.refresh = refreshLastAccessOnGet

	go func() {
		ticker := time.NewTicker(pruneInterval)
		defer ticker.Stop() // Clean up the ticker when goroutine exits

		for {
			select {
			case <-m.done:
				ticker.Stop()
				return
			case now := <-ticker.C:
				m.l.Lock()
				for k, v := range m.m {
					if now.Sub(v.lastAccess) >= maxTTL {
						delete(m.m, k)
					}
				}
				m.l.Unlock()
			}
		}
	}()
	return
}

func (m *TtlMap[T]) Len() int {
	m.l.Lock()
	defer m.l.Unlock()
	return len(m.m)
}

func (m *TtlMap[T]) Put(k T, v interface{}) {
	m.l.Lock()
	m.m[k] = &item{Value: v, lastAccess: time.Now()}
	m.l.Unlock()
}

func (m *TtlMap[T]) Get(k T) (v interface{}) {
	m.l.Lock()
	if it, ok := m.m[k]; ok {
		v = it.Value
		if m.refresh {
			m.m[k].lastAccess = time.Now()
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

// All copies the entries, but does not deep-copy their Value payloads.
func (m *TtlMap[T]) All() map[T]*item {
	m.l.Lock()
	dst := make(map[T]*item, len(m.m))
	for k, v := range m.m {
		copied := *v
		dst[k] = &copied
	}
	m.l.Unlock()
	return dst
}

func (m *TtlMap[T]) Close() {
	m.closeOnce.Do(func() {
		close(m.done)
	})
}
