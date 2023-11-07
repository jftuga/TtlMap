package TtlMap

import (
	"maps"
	"testing"
	"time"
)

func TestAllItemsExpired(t *testing.T) {
	maxTTL := time.Duration(time.Second * 4)        // time in seconds
	startSize := 3                                  // initial number of items in map
	pruneInterval := time.Duration(time.Second * 1) // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true                  // update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int_array", []int{1, 2, 3})

	time.Sleep(maxTTL + pruneInterval)
	t.Logf("tm.len: %v\n", tm.Len())
	if tm.Len() > 0 {
		t.Errorf("t.Len should be 0, but actually equals %v\n", tm.Len())
	}
}

func TestNoItemsExpired(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)        // time in seconds
	startSize := 3                                  // initial number of items in map
	pruneInterval := time.Duration(time.Second * 3) // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true                  // update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int_array", []int{1, 2, 3})

	time.Sleep(maxTTL)
	t.Logf("tm.len: %v\n", tm.Len())
	if tm.Len() != 2 {
		t.Fatalf("t.Len should equal 2, but actually equals %v\n", tm.Len())
	}
}

func TestKeepFloat(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)        // time in seconds
	startSize := 3                                  // initial number of items in map
	pruneInterval := time.Duration(time.Second * 1) // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true                  // update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int", 1234)
	tm.Put("int_array", []int{1, 2, 3})

	dontExpireKey := "int"
	go func() {
		for range time.Tick(time.Second) {
			tm.Get(CustomKeyType(dontExpireKey))
		}
	}()

	time.Sleep(maxTTL + pruneInterval)
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
	maxTTL := time.Duration(time.Second * 4)        // time in seconds
	startSize := 3                                  // initial number of items in map
	pruneInterval := time.Duration(time.Second * 1) // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := false                 // do NOT update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int_array", []int{1, 2, 3})

	go func() {
		for range time.Tick(time.Second) {
			tm.Get("myString")
			tm.Get("int_array")
		}
	}()

	time.Sleep(maxTTL + pruneInterval)
	t.Logf("tm.Len: %v\n", tm.Len())
	if tm.Len() != 0 {
		t.Errorf("t.Len should be 0, but actually equals %v\n", tm.Len())
	}
}

func TestDelete(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)        // time in seconds
	startSize := 3                                  // initial number of items in map
	pruneInterval := time.Duration(time.Second * 4) // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true                  // update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// populate the TtlMap
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
}

func TestClear(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)        // time in seconds
	startSize := 3                                  // initial number of items in map
	pruneInterval := time.Duration(time.Second * 4) // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true                  // update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// populate the TtlMap
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

func TestAllFunc(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)        // time in seconds
	startSize := 3                                  // initial number of items in map
	pruneInterval := time.Duration(time.Second * 4) // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true                  // update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int", 1234)
	tm.Put("floatPi", 3.1415)
	tm.Put("int_array", []int{1, 2, 3})
	tm.Put("boolean", true)

	tm.Delete("floatPi")
	//t.Logf("tm.len: %v\n", tm.Len())
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

func TestGetNoUpdate(t *testing.T) {
	maxTTL := time.Duration(time.Second * 2)        // time in seconds
	startSize := 3                                  // initial number of items in map
	pruneInterval := time.Duration(time.Second * 4) // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true                  // update item's lastAccessTime on a .Get()
	tm := New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer tm.Close()

	// populate the TtlMap
	tm.Put("myString", "a b c")
	tm.Put("int", 1234)
	tm.Put("floatPi", 3.1415)
	tm.Put("int_array", []int{1, 2, 3})
	tm.Put("boolean", true)

	go func() {
		for range time.Tick(time.Second) {
			tm.GetNoUpdate("myString")
			tm.GetNoUpdate("int_array")
		}
	}()

	time.Sleep(maxTTL + pruneInterval)
	t.Logf("tm.Len: %v\n", tm.Len())
	if tm.Len() != 0 {
		t.Errorf("t.Len should be 0, but actually equals %v\n", tm.Len())
	}
}
