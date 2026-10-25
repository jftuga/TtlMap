package TtlMap

import (
	"sync"
	"testing"
	"time"
)

func TestTTLDoesNotExpireEarly(t *testing.T) {
	for _, ttl := range []time.Duration{500 * time.Millisecond, time.Second} {
		t.Run(ttl.String(), func(t *testing.T) {
			m := New[string](ttl, 1, 5*time.Millisecond, false)
			defer m.Close()
			start := time.Now()
			m.Put("key", "value")

			// Observe the entire lifetime, including any wall-clock second boundary.
			deadline := start.Add(ttl + time.Second)
			for {
				value := m.GetNoUpdate("key")
				elapsed := time.Since(start)
				if value == nil {
					if elapsed < ttl {
						t.Fatalf("TTL %v expired after only %v", ttl, elapsed)
					}
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("entry did not expire within %v", ttl+time.Second)
				}
				time.Sleep(time.Millisecond)
			}
		})
	}
}

func TestConcurrentLen(t *testing.T) {
	m := New[int](time.Hour, 1, time.Hour, false)
	defer m.Close()
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 10000; i++ {
			m.Put(i, i)
			m.Delete(i)
		}
	}()
	close(start)
	for i := 0; i < 10000; i++ {
		if n := m.Len(); n < 0 || n > 1 {
			t.Errorf("unexpected length: %d", n)
			break
		}
	}
	wg.Wait()
	if n := m.Len(); n != 0 {
		t.Fatalf("expected empty map, got length %d", n)
	}
}

func TestConcurrentClose(t *testing.T) {
	for round := 0; round < 500; round++ {
		m := New[int](time.Hour, 1, time.Hour, false)
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i := 0; i < 32; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				m.Close()
			}()
		}
		close(start)
		wg.Wait()
		m.Close()
		select {
		case <-m.done:
		default:
			t.Fatal("Close did not close the shutdown channel")
		}
	}
}
