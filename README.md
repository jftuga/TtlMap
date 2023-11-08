# TtlMap

`TtlMap` is golang package that implements a *time-to-live* map such that after a given amount of time, items in the map are deleted.
* The map key can be any [comparable](https://go.dev/ref/spec#Comparison_operators) data type, via Generics.
* Any data type can be used as a map value. Internally, `interface{}` is used for this.

## Example

[Full example using many data types](example/example.go)

Small example:

```go
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
	t := TtlMap.New[string](maxTTL, startSize, pruneInterval, refreshLastAccessOnGet)
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
```

Output:

```
$ go run small.go

TtlMap length: 2
[ myString] a b c
[int_array] [1 2 3]

Sleeping 5 seconds, items should be 'nil' after this time
[ myString] <nil>
[int_array] <nil>
TtlMap length: 0
```

## API functions
* `New`: initialize a `TtlMap`
* `Close`: this stops the goroutine that checks for expired items; use with `defer`
* `Len`: return the number of items in the map
* `Put`: add a key/value
* `Get`: get the current value of the given key; return `nil` if the key is not found in the map
* `GetNoUpdate`: same as `Get`, but do not update the `lastAccess` expiration time
* * * This ignores the `refreshLastAccessOnGet` parameter
* `All`: returns a *copy* of all items in the map
* `Delete`: delete an item; return `true` if the item was deleted, `false` if the item was not found in the map
* `Clear`: remove all items from the map

## Performance
* Searching for expired items runs in O(n) time, where n = number of items in the `TtlMap`.
* * This inefficiency can be somewhat mitigated by increasing the value of the `pruneInterval` time.
* In most cases you want `pruneInterval > maxTTL`; otherwise expired items will stay in the map longer than expected.

## Acknowledgments
* Adopted from: [Map with TTL option in Go](https://stackoverflow.com/a/25487392/452281)
* * Answer created by: [OneOfOne](https://stackoverflow.com/users/145587/oneofone)
* [/u/skeeto](https://old.reddit.com/user/skeeto): suggestions for the `New` function

## Disclosure Notification

This program was completely developed on my own personal time, for my own personal benefit, and on my personally owned equipment.
