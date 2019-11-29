package hashtable_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/spaolacci/murmur3"
	"github.com/yeqown/hashtable"
)

func Test_murmur3x128(t *testing.T) {
	hasher := murmur3.New128()
	hasher.Write([]byte("this is key"))
	v1, v2 := hasher.Sum128()
	t.Log(v1, v2)

	hasher.Reset()
	hasher.Write([]byte("this is key2"))
	v1, v2 = hasher.Sum128()
	t.Log(v1, v2)
}

func Test_murmur3x64(t *testing.T) {
	hasher := murmur3.New64()
	hasher.Write([]byte("this is key"))
	v1 := hasher.Sum64()
	t.Log(v1)

	hasher.Reset()
	hasher.Write([]byte("this is key2"))
	v1 = hasher.Sum64()
	t.Log(v1)
}

func Test_LinkedDict_SetGetDel(t *testing.T) {
	m := hashtable.NewLinkedDict()

	m.Set("author", "yeqown")
	if v, ok := m.Get("author"); !ok || v.(string) != "yeqown" {
		t.Errorf("want got=true, v=yeqown, actual got=%v, v=%v", ok, v)
		t.FailNow()
	}

	m.Set("date", "20191128")
	if v, ok := m.Get("date"); !ok || v.(string) != "20191128" {
		t.Errorf("want got=true, v=20191128, actual got=%v, v=%v", ok, v)
		t.FailNow()
	}

	m.Set("date", "2019/11/28 03:50PM")
	if v, ok := m.Get("date"); !ok || v.(string) != "2019/11/28 03:50PM" {
		t.Errorf("want got=true, v='2019/11/28 03:50PM', actual got=%v, v=%v", ok, v)
		t.FailNow()
	}

	m.Del("date")
	if v, ok := m.Get("date"); ok || v != nil {
		t.Errorf("want got=false, v=nil, actual got=%v, v=%v", ok, v)
		t.FailNow()
	}

	m.Del("author")
	if v, ok := m.Get("author"); ok || v != nil {
		t.Errorf("want got=false, v=nil, actual got=%v, v=%v", ok, v)
		t.FailNow()
	}

	if size := m.Len(); size != 0 {
		t.Errorf("want Len()=0,actual got=%d", size)
		t.FailNow()
	}
}

func Test_LinkedDict_Iter(t *testing.T) {
	m := hashtable.NewLinkedDict()
	for i := 0; i < 100000; i++ {
		m.Set(fmt.Sprintf("key_%d", i), i)
	}

	cache := make(map[string]bool)
	m.Iter(func(key string, v interface{}) {
		cache[key] = true
	})

	if size := len(cache); size != 100000 {
		t.Errorf("invalid iter func, want=100000, got cache=%d, dictSize=%d", size, m.Len())
		t.FailNow()
	}
}

func Test_LinkedDict_Rehash(t *testing.T) {
	// bydebug to test
	m := hashtable.NewLinkedDict()
	for i := 0; i < 1024; i++ {
		m.Set(fmt.Sprintf("key_%d", i), i)
	}
	t.Log("finished")
}
func Test_LinkedDict_Shrink(t *testing.T) {
	m := hashtable.NewLinkedDict()
	for i := 0; i < 1024; i++ {
		key := fmt.Sprintf("key_%d", i)
		m.Set(key, i)
		m.Get(key)
	}

	t.Log(m.Len())

	for i := 0; i < 1000; i++ {
		if i > 900 {
			t.Log(m.Len())
		}
		m.Del(fmt.Sprintf("key_%d", i))
	}
}

func Benchmark_LinkedDict(b *testing.B) {
	/*
		goos: darwin
		goarch: amd64
		pkg: github.com/yeqown/hashtable
		Benchmark_LinkedDict-4   	 4458631	       268 ns/op	      23 B/op	       2 allocs/op
		PASS
		ok  	github.com/yeqown/hashtable	1.769s
		Success: Benchmarks passed.
	*/
	var (
		key string
		v   interface{}
		ok  bool
	)
	b.StopTimer()
	m := hashtable.NewLinkedDict()
	for i := 0; i < 100000; i++ {
		key = fmt.Sprintf("key_%d", i)
		m.Set(key, i)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		pos := rand.Intn(100000)
		key = fmt.Sprintf("key_%d", pos)
		if v, ok = m.Get(key); !ok || v.(int) != pos {
			b.Errorf("want got=true, v=%d, actual got=%v, v=%v", pos, ok, v)
			b.FailNow()
		}
	}
}

func XBenchmark_goMap(b *testing.B) {
	/*
		goos: darwin
		goarch: amd64
		pkg: github.com/yeqown/hashtable
		Benchmark_goMap-4   	 4861671	       230 ns/op	      15 B/op	       1 allocs/op
		PASS
		ok  	github.com/yeqown/hashtable	1.869s
		Success: Benchmarks passed.
	*/
	b.StopTimer()
	m := make(map[string]interface{})
	for i := 0; i < 1024; i++ {
		key := fmt.Sprintf("key_%d", i)
		m[key] = i
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		pos := rand.Intn(1024)
		if v, ok := m[fmt.Sprintf("key_%d", pos)]; !ok || v.(int) != pos {
			b.Errorf("want got=true, v=%d, actual got=%v, v=%v", pos, ok, v)
			b.FailNow()
		}
	}
}

// condition: used=1024 size=4，没有自动扩容没有rehash
//
// goos: darwin
// goarch: amd64
// pkg: github.com/yeqown/hashtable
// Benchmark_LinkedDict-4   	 1000000	      1021 ns/op	      23 B/op	       2 allocs/op
// PASS
// ok  	github.com/yeqown/hashtable	1.373s

// condition: used=1024, size=1024, 有自动扩容
//
// goos: darwin
// goarch: amd64
// pkg: github.com/yeqown/hashtable
// Benchmark_LinkedDict-4   	 5000000	       256 ns/op	      23 B/op	       2 allocs/op
// PASS
// ok  	github.com/yeqown/hashtable	1.918s

// conditions: used=1024 size=1024, 有自动扩容和自动缩容
//
// goos: windows
// goarch: amd64
// pkg: github.com/yeqown/hashtable
// Benchmark_LinkedDict-4   	10000000	       244 ns/op	      23 B/op	       2 allocs/op
// PASS
// ok  	github.com/yeqown/hashtable	2.888s
// Success: Benchmarks passed.
