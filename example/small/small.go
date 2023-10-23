package main

import (
	"fmt"
	"time"

	"github.com/jftuga/ttlMap"
)

func main() {
	maxTTL := 4                    // time in seconds
	startSize := 3                 // initial number of items in map
	pruneInterval := 1             // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true // update item's lastAccessTime on a .Get()
	t := ttlMap.New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)

	// populate the ttlMap
	t.Put("myString", "a b c")
	t.Put("int_array", []int{1, 2, 3})
	fmt.Println("ttlMap length:", t.Len())

	// display all items in ttlMap
	all := t.All()
	for k, v := range all {
		fmt.Printf("[%9s] %v\n", k, v.Value)
	}
	fmt.Println()

	sleepTime := maxTTL + pruneInterval
	fmt.Printf("Sleeping %v seconds, items should be 'nil' after this time\n", sleepTime)
	time.Sleep(time.Second * time.Duration(sleepTime))
	fmt.Printf("[%9s] %v\n", "myString", t.Get("myString"))
	fmt.Printf("[%9s] %v\n", "int_array", t.Get("int_array"))
	fmt.Println("ttlMap length:", t.Len())
}
