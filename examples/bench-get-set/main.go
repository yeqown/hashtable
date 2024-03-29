package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/yeqown/hashtable"
)

func main() {
	var N = 1000
	// 1K
	testBuiltinMap(N)
	testMap(N)

	// 10K
	N = 10000
	testBuiltinMap(N)
	testMap(N)

	// 100K
	N = 100000
	testBuiltinMap(N)
	testMap(N)

	// 1M
	N = 1000000
	testBuiltinMap(N)
	testMap(N)

	// 10M
	N = 10000000
	testBuiltinMap(N)
	testMap(N)
}

func testBuiltinMap(N int) {
	var (
		builtCost int64
		getCost   int64
	)

	var (
		m   = make(map[string]interface{})
		key string
		v   interface{}
		ok  bool
	)

	startSet := time.Now()
	for i := 0; i < N; i++ {
		key = fmt.Sprintf("key_%d", i)
		m[key] = i
	}
	builtCost = time.Now().Sub(startSet).Milliseconds()
	fmt.Printf("builtinMap_%d, cost: %d\n", N, builtCost)

	startGet := time.Now()
	for i := 0; i < N; i++ {
		pos := rand.Intn(N)
		key = fmt.Sprintf("key_%d", pos)
		// fmt.Print(key)
		v, ok = m[key]
		_, _ = v, ok
	}
	getCost = time.Now().Sub(startGet).Milliseconds()
	fmt.Printf("getMap_%d, cost: %d\n", N, getCost)

}

func testMap(N int) {
	var (
		builtCost int64
		getCost   int64
	)

	var (
		m   = hashtable.New[string, int]()
		key string
		v   interface{}
		ok  bool
	)

	startSet := time.Now()
	for i := 0; i < N; i++ {
		key = fmt.Sprintf("key_%d", i)
		m.Set(key, i)
	}
	builtCost = time.Now().Sub(startSet).Milliseconds()
	fmt.Printf("builtinMap_%d, cost: %d\n", N, builtCost)

	startGet := time.Now()
	for i := 0; i < N; i++ {
		pos := rand.Intn(N)
		key = fmt.Sprintf("key_%d", pos)
		v, ok = m.Lookup(key)
		_, _ = v, ok
	}
	getCost = time.Now().Sub(startGet).Milliseconds()
	fmt.Printf("getMap_%d, cost: %d\n", N, getCost)
}
