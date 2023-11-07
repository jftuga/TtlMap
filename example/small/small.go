package main

import (
	"fmt"
	"time"

	"github.com/jftuga/TtlMap"
)

func main() {
	maxTTL := time.Duration(time.Second * 4)        // a key's time to live in seconds
	startSize := 3                                  // initial number of items in map
	pruneInterval := time.Duration(time.Second * 1) // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true                  // update item's 'lastAccessTime' on a .Get()
	t := TtlMap.New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
	defer t.Close()

	// populate the TtlMap
	t.Put("myString", "a b c")
	t.Put("int_array", []int{1, 2, 3})
	fmt.Println("TtlMap length:", t.Len())

	// display all items in TtlMap
	all := t.All()
	for k, v := range all {
		fmt.Printf("[%9s] %v\n", k, v.Value)
	}
	fmt.Println()

	sleepTime := maxTTL + pruneInterval
	fmt.Printf("Sleeping %v seconds, items should be 'nil' after this time\n", sleepTime)
	time.Sleep(sleepTime)
	fmt.Printf("[%9s] %v\n", "myString", t.Get("myString"))
	fmt.Printf("[%9s] %v\n", "int_array", t.Get("int_array"))
	fmt.Println("TtlMap length:", t.Len())
}
