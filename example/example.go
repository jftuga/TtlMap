/*
example.go
-John Taylor
2023-10-21

This is an example on how to use the ttlMap package.  Notice the variety of data types used.
*/

package main

import (
	"fmt"
	"time"

	"github.com/jftuga/ttlMap"
)

type User struct {
	Name  string
	Level uint
}

func main() {
	maxTTL := 4                    // time in seconds
	startSize := 3                 // initial number of items in map
	pruneInterval := 1             // search for expired items every 'pruneInterval' seconds
	refreshLastAccessOnGet := true // update item's lastAccessTime on a .Get()
	t := ttlMap.New(maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)

	// populate the ttlMap
	t.Put("string", "a b c")
	t.Put("int", 3)
	t.Put("float", 4.4)
	t.Put("int_array", []int{1, 2, 3})
	t.Put("bool", false)
	t.Put("rune", '{')
	t.Put("byte", 0x7b)
	var u = uint64(123456789)
	t.Put("uint64", u)
	var c = complex(3.14, -4.321)
	t.Put("complex", c)

	allUsers := []User{{Name: "abc", Level: 123}, {Name: "def", Level: 456}}
	t.Put("all_users", allUsers)

	fmt.Println()
	fmt.Println("ttlMap length:", t.Len())

	// extract entry from struct array
	a := t.Get("all_users").([]User)
	fmt.Printf("second user: %v, %v\n", a[1].Name, a[1].Level)

	// display all items in ttlMap
	fmt.Println()
	fmt.Println("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	all := t.All()
	for k, v := range all {
		fmt.Printf("[%9s] %v\n", k, v.Value)
	}
	fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
	fmt.Println()

	// by executing Get(), the 'dontExpireKey' lastAccessTime will be updated
	// therefore, this item will not expire
	dontExpireKey := "float"
	go func() {
		for range time.Tick(time.Second) {
			t.Get(ttlMap.CustomKeyType(dontExpireKey))
		}
	}()

	// ttlMap has an expiration time, wait until this amount of time passes
	sleepTime := maxTTL + pruneInterval
	fmt.Println()
	fmt.Printf("Sleeping %v seconds, items should be removed after this time, except for the '%v' key\n", sleepTime, dontExpireKey)
	fmt.Println()
	time.Sleep(time.Second * time.Duration(sleepTime))

	// these items have expired and therefore should be nil, except for 'dontExpireKey'
	fmt.Printf("[%9s] %v\n", "string", t.Get("string"))
	fmt.Printf("[%9s] %v\n", "int", t.Get("int"))
	fmt.Printf("[%9s] %v\n", "float", t.Get("float"))
	fmt.Printf("[%9s] %v\n", "int_array", t.Get("int_array"))
	fmt.Printf("[%9s] %v\n", "bool", t.Get("bool"))
	fmt.Printf("[%9s] %v\n", "rune", t.Get("rune"))
	fmt.Printf("[%9s] %v\n", "byte", t.Get("byte"))
	fmt.Printf("[%9s] %v\n", "uint64", t.Get("uint64"))
	fmt.Printf("[%9s] %v\n", "complex", t.Get("complex"))
	fmt.Printf("[%9s] %v\n", "all_users", t.Get("all_users"))

	// sanity check, this comparison should be true
	fmt.Println()
	if t.Get("int") == nil {
		fmt.Println("[int] is nil")
	}
	fmt.Println("ttlMap length:", t.Len())
	fmt.Println()

	fmt.Println()
	fmt.Printf("Manually deleting '%v' key; should be successful\n", dontExpireKey)
	success := t.Delete(ttlMap.CustomKeyType(dontExpireKey))
	fmt.Printf("    successful? %v\n", success)
	fmt.Printf("Manually deleting '%v' key again; should NOT be successful this time\n", dontExpireKey)
	success = t.Delete(ttlMap.CustomKeyType(dontExpireKey))
	fmt.Printf("    successful? %v\n", success)
	fmt.Println("ttlMap length:", t.Len())
	fmt.Println()
}
