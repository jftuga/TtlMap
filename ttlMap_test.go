/*
ttlMap_test.go

Comprehensive test suite for TtlMap that combines memory leak detection,
goroutine tracking, benchmarks, and functional tests for all TtlMap operations.
*/

package TtlMap

import (
	"maps"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"
)

// ====================
// Memory Leak Detection Tests
// ====================

// Test 1: Simple Memory Stats Test
// Quick test to detect substantial memory growth
func TestMemoryLeak(t *testing.T) {
	var m runtime.MemStats

	// Force GC and get baseline
	runtime.GC()
	runtime.ReadMemStats(&m)
	baseline := m.Alloc

	// Create and close 1000 TtlMaps
	for i := 0; i < 1000; i++ {
		ttlMap := New[string](5*time.Second, 10, 1*time.Second, true)
		ttlMap.Put("key", "value")
		ttlMap.Close()
	}

	// Force GC and check memory
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // Give GC time to clean up
	runtime.ReadMemStats(&m)
	after := m.Alloc

	growth := after - baseline
	t.Logf("Memory baseline: %d bytes", baseline)
	t.Logf("Memory after: %d bytes", after)
	t.Logf("Memory growth: %d bytes", growth)

	// If there's a leak, growth will be substantial (millions of bytes)
	// Without leak, should be minimal (< 1 MB)
	if growth > 1024*1024 {
		t.Errorf("Potential memory leak detected: %d bytes growth", growth)
	}
}

// Test 2: Memory Profile with pprof
// Creates before/after memory profiles for detailed analysis
func TestMemoryLeakProfile(t *testing.T) {
	// Create profile file
	f, err := os.Create("mem_before.prof")
	if err != nil {
		t.Fatal(err)
	}
	runtime.GC()
	pprof.WriteHeapProfile(f)
	f.Close()

	// Create and close many TtlMaps
	for i := 0; i < 10000; i++ {
		ttlMap := New[string](5*time.Second, 10, 1*time.Second, true)
		ttlMap.Put("key", "value")
		ttlMap.Close()
	}

	// Create after profile
	f2, err := os.Create("mem_after.prof")
	if err != nil {
		t.Fatal(err)
	}
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	pprof.WriteHeapProfile(f2)
	f2.Close()

	t.Log("Memory profiles created: mem_before.prof and mem_after.prof")
	t.Log("Run: go tool pprof -base=mem_before.prof mem_after.prof")
}

// Test 3: Goroutine Leak Check
// Verifies that goroutines are properly cleaned up after Close()
func TestGoroutineLeak(t *testing.T) {
	baseline := runtime.NumGoroutine()

	// Create and close 100 TtlMaps
	for i := 0; i < 100; i++ {
		ttlMap := New[string](5*time.Second, 10, 1*time.Second, true)
		ttlMap.Close()
	}

	time.Sleep(100 * time.Millisecond) // Give goroutines time to exit

	after := runtime.NumGoroutine()
	leaked := after - baseline

	t.Logf("Goroutines at start: %d", baseline)
	t.Logf("Goroutines after: %d", after)
	t.Logf("Leaked goroutines: %d", leaked)

	// Should be 0 or very close (test goroutine itself may vary)
	if leaked > 5 {
		t.Errorf("Goroutine leak detected: %d goroutines not cleaned up", leaked)
	}
}

// Test 4: Benchmark with Memory Allocation
// Measures performance and memory allocation per operation
func BenchmarkTtlMapCreateDestroy(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ttlMap := New[string](5*time.Second, 10, 100*time.Millisecond, true)
		ttlMap.Put("key", "value")
		ttlMap.Close()
	}
}

// ====================
// Basic Functionality Tests
// ====================

// Verify basic Put, Get, Delete, and TTL expiration
func TestBasicFunctionality(t *testing.T) {
	ttlMap := New[string](2*time.Second, 10, 500*time.Millisecond, true)
	defer ttlMap.Close()

	// Test Put and Get
	ttlMap.Put("key1", "value1")
	val := ttlMap.Get("key1")
	if val != "value1" {
		t.Errorf("Expected 'value1', got '%v'", val)
	}

	// Test Len
	if ttlMap.Len() != 1 {
		t.Errorf("Expected length 1, got %d", ttlMap.Len())
	}

	// Test Delete
	if !ttlMap.Delete("key1") {
		t.Error("Expected Delete to return true")
	}
	if ttlMap.Len() != 0 {
		t.Errorf("Expected length 0 after delete, got %d", ttlMap.Len())
	}

	// Test TTL expiration
	ttlMap.Put("key2", "value2")
	time.Sleep(2500 * time.Millisecond) // Wait for TTL + pruneInterval
	val = ttlMap.Get("key2")
	if val != nil {
		t.Errorf("Expected key2 to be expired, but got '%v'", val)
	}
}

// Test refresh functionality with Get() vs GetNoUpdate()
func TestRefreshOnGet(t *testing.T) {
	// Test with refresh enabled - use longer TTL for more reliable timing
	ttlMap := New[int](5*time.Second, 10, 500*time.Millisecond, true)
	defer ttlMap.Close()

	ttlMap.Put(1, "value1")
	time.Sleep(2 * time.Second)
	ttlMap.Get(1)               // Should refresh lastAccess
	time.Sleep(3 * time.Second) // Total: 5s, but lastAccess was refreshed at 2s
	val := ttlMap.Get(1)
	if val == nil {
		t.Error("Expected key to still exist due to refresh on Get")
	}

	// Test GetNoUpdate - use fresh key with more margin
	ttlMap.Put(2, "value2")
	time.Sleep(2 * time.Second)
	ttlMap.GetNoUpdate(2)               // Should NOT refresh lastAccess
	time.Sleep(3500 * time.Millisecond) // Total: 5.5s > 5s TTL
	val = ttlMap.Get(2)
	if val != nil {
		t.Error("Expected key to be expired since GetNoUpdate doesn't refresh")
	}
}

// ====================
// Additional Functional Tests
// ====================

// Test Clear() method
func TestClear(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)
	startSize := 3
	pruneInterval := time.Duration(time.Second * 4)
	refreshLastAccessOnGet := true
	tm := New[string](maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// Populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int_array", []int{1, 2, 3})
	t.Logf("tm.len: %v\n", tm.Len())
	if tm.Len() != 2 {
		t.Fatalf("t.Len should equal 2, but actually equals %v\n", tm.Len())
	}

	tm.Clear()
	t.Logf("tm.len: %v\n", tm.Len())
	if tm.Len() != 0 {
		t.Fatalf("t.Len should equal 0, but actually equals %v\n", tm.Len())
	}
}

// Test All() method returns a proper copy
func TestAllFunc(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)
	startSize := 3
	pruneInterval := time.Duration(time.Second * 4)
	refreshLastAccessOnGet := true
	tm := New[string](maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// Populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int", 1234)
	tm.Put("floatPi", 3.1415)
	tm.Put("int_array", []int{1, 2, 3})
	tm.Put("boolean", true)

	tm.Delete("floatPi")
	if tm.Len() != 4 {
		t.Fatalf("t.Len should equal 4, but actually equals %v\n", tm.Len())
	}

	tm.Put("byte", 0x7b)
	var u = uint64(123456789)
	tm.Put("uint64", u)

	allItems := tm.All()
	if !maps.Equal(allItems, tm.m) {
		t.Fatalf("allItems and tm.m are not equal\n")
	}
}

// Test multiple Put() operations on the same key
func TestMultiplePuts(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)
	startSize := 3
	pruneInterval := time.Duration(time.Second * 4)
	refreshLastAccessOnGet := true
	tm := New[string](maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	key := "example"
	tm.Put(key, "original")

	tm.Put(key, "revised")
	if tm.Get(key) != "revised" {
		t.Errorf("The '%v' should equal 'revised', but actually equals: '%v'\n", key, tm.Get(key))
	}

	tm.Put(key, "revised-2")
	if tm.Get(key) != "revised-2" {
		t.Errorf("The '%v' should equal 'revised-2', but actually equals: '%v'\n", key, tm.Get(key))
	}
}

// Test with uint64 keys to verify generic key types work
func TestUInt64Key(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)
	startSize := 3
	pruneInterval := time.Duration(time.Second * 4)
	refreshLastAccessOnGet := true
	tm := New[uint64](maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	tm.Put(18446744073709551615, "largest")
	tm.Put(9223372036854776000, "mid")
	tm.Put(0, "zero")

	allItems := tm.All()
	for k, v := range allItems {
		t.Logf("k: %v   v: %v\n", k, v.Value)
	}

	time.Sleep(maxTTL + pruneInterval)
	t.Logf("tm.Len: %v\n", tm.Len())
	if tm.Len() != 0 {
		t.Errorf("t.Len should be 0, but actually equals %v\n", tm.Len())
	}
}

// Test keeping one item alive via periodic Get() while others expire
// This version properly cleans up the refresh goroutine to avoid leaks
func TestKeepItemAlive(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)
	startSize := 3
	pruneInterval := time.Duration(time.Second * 1)
	refreshLastAccessOnGet := true
	tm := New[string](maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// Populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int", 1234)
	tm.Put("int_array", []int{1, 2, 3})

	dontExpireKey := "int"

	// Create a ticker that we can stop
	ticker := time.NewTicker(time.Second)
	done := make(chan bool)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				tm.Get(dontExpireKey)
			}
		}
	}()

	time.Sleep(maxTTL + pruneInterval)

	// Stop the refresh goroutine
	close(done)
	time.Sleep(100 * time.Millisecond) // Give goroutine time to exit

	if tm.Len() != 1 {
		t.Fatalf("t.Len should equal 1, but actually equals %v\n", tm.Len())
	}
	all := tm.All()
	if all[dontExpireKey].Value != 1234 {
		t.Errorf("Value should equal 1234 but actually equals %v\n", all[dontExpireKey].Value)
	}
	t.Logf("tm.Len: %v\n", tm.Len())
	t.Logf("%v Value: %v\n", dontExpireKey, all[dontExpireKey].Value)
}

// Test that with refreshLastAccessOnGet=false, items expire even if accessed
func TestWithNoRefresh(t *testing.T) {
	maxTTL := time.Duration(time.Second * 4)
	startSize := 3
	pruneInterval := time.Duration(time.Second * 1)
	refreshLastAccessOnGet := false // Do NOT update item's lastAccessTime on a .Get()
	tm := New[string](maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// Populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int_array", []int{1, 2, 3})

	// Create a ticker that we can stop
	ticker := time.NewTicker(time.Second)
	done := make(chan bool)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				tm.Get("myString")
				tm.Get("int_array")
			}
		}
	}()

	time.Sleep(maxTTL + pruneInterval)

	// Stop the refresh goroutine
	close(done)
	time.Sleep(100 * time.Millisecond) // Give goroutine time to exit

	t.Logf("tm.Len: %v\n", tm.Len())
	if tm.Len() != 0 {
		t.Errorf("t.Len should be 0, but actually equals %v\n", tm.Len())
	}
}

// Test Delete() method returns proper boolean values
func TestDelete(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)
	startSize := 3
	pruneInterval := time.Duration(time.Second * 4)
	refreshLastAccessOnGet := true
	tm := New[string](maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// Populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int_array", []int{1, 2, 3})

	tm.Delete("int_array")
	t.Logf("tm.len: %v\n", tm.Len())
	if tm.Len() != 1 {
		t.Fatalf("t.Len should equal 1, but actually equals %v\n", tm.Len())
	}

	tm.Delete("myString")
	t.Logf("tm.len: %v\n", tm.Len())
	if tm.Len() != 0 {
		t.Fatalf("t.Len should equal 0, but actually equals %v\n", tm.Len())
	}

	// Test deleting non-existent key
	if tm.Delete("nonExistent") {
		t.Error("Delete should return false for non-existent key")
	}
}
