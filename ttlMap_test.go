package ttlMap

import (
	"testing"
	"time"
)

func TestAllItemsExpired(t *testing.T) {
	maxTTL := 4                    // time in seconds
	startSize := 3                 // initial number of items in map
	pruneInterval := 1             // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true // update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)

	// populate the ttlMap
	tm.Put("myString", "a b c")
	tm.Put("int_array", []int{1, 2, 3})

	sleepTime := maxTTL + pruneInterval
	time.Sleep(time.Second * time.Duration(sleepTime))
	t.Logf("tm.len: %v\n", tm.Len())
	if tm.Len() > 0 {
		t.Errorf("t.Len should be 0, but actually equals %v\n", tm.Len())
	}
}

func TestNoItemsExpired(t *testing.T) {
	maxTTL := 2                    // time in seconds
	startSize := 3                 // initial number of items in map
	pruneInterval := 3             // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true // update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)

	// populate the ttlMap
	tm.Put("myString", "a b c")
	tm.Put("int_array", []int{1, 2, 3})

	sleepTime := maxTTL
	time.Sleep(time.Second * time.Duration(sleepTime))
	t.Logf("tm.len: %v\n", tm.Len())
	if tm.Len() != 2 {
		t.Fatalf("t.Len should equal 2, but actually equals %v\n", tm.Len())
	}
}

func TestKeepFloat(t *testing.T) {
	maxTTL := 2                    // time in seconds
	startSize := 3                 // initial number of items in map
	pruneInterval := 1             // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true // update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)

	// populate the ttlMap
	tm.Put("myString", "a b c")
	tm.Put("int", 1234)
	tm.Put("int_array", []int{1, 2, 3})

	dontExpireKey := "int"
	go func() {
		for range time.Tick(time.Second) {
			tm.Get(CustomKeyType(dontExpireKey))
		}
	}()

	sleepTime := maxTTL + pruneInterval
	time.Sleep(time.Second * time.Duration(sleepTime))
	if tm.Len() != 1 {
		t.Fatalf("t.Len should equal 1, but actually equals %v\n", tm.Len())
	}
	all := tm.All()
	if all[CustomKeyType(dontExpireKey)].Value != 1234 {
		t.Errorf("Value should equal 1234 but actually equals %v\n", all[CustomKeyType(dontExpireKey)].Value)
	}
	t.Logf("tm.Len: %v\n", tm.Len())
	t.Logf("%v Value: %v\n", dontExpireKey, all[CustomKeyType(dontExpireKey)].Value)
}

func TestWithNoRefresh(t *testing.T) {
	maxTTL := 4                     // time in seconds
	startSize := 3                  // initial number of items in map
	pruneInterval := 1              // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := false // do NOT update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)

	// populate the ttlMap
	tm.Put("myString", "a b c")
	tm.Put("int_array", []int{1, 2, 3})

	go func() {
		for range time.Tick(time.Second) {
			tm.Get("myString")
			tm.Get("int_array")
		}
	}()

	sleepTime := maxTTL + pruneInterval
	time.Sleep(time.Second * time.Duration(sleepTime))
	t.Logf("tm.Len: %v\n", tm.Len())
	if tm.Len() != 0 {
		t.Errorf("t.Len should be 0, but actually equals %v\n", tm.Len())
	}
}