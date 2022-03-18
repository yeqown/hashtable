package hashtable

import (
	"github.com/spaolacci/murmur3"
	"unsafe"
)

type _rehashEnum uint8

const (
	// _GROWUP indicates that the map is growing up.
	_GROWUP _rehashEnum = iota + 1
	// _SHRINK indicate the map is step shrink.
	_SHRINK
)

const (
	// _INIT_TABLE_CAP means the initial table capacity.
	_INIT_TABLE_CAP = 4
	// _REHASH_GROW_THRESHOLD rehash grow threshold of the map size factor.
	_REHASH_GROW_THRESHOLD = 1.0
	// _REHASH_SHRINK_THRESHOLD represents the threshold of the map size factor
	// when the table will be shrinked.
	_REHASH_SHRINK_THRESHOLD = 10
)

// Map is hashtable implemented based on redis hashtable in go,
// it resolves collision by chaining.
type Map[K comparable, V any] struct {
	// two hashtable to work with, ht[0] will be use normally,
	// ht[1] will only be used when rehashing
	ht [2]*hashtable[K, V]

	// rehashing flag, -1 means not in rehashing
	rehashIdx int
}

// New to create a Map which is implemented based on linked-list
func New[K comparable, V any]() *Map[K, V] {
	m := &Map[K, V]{
		ht:        [2]*hashtable[K, V]{nil, nil},
		rehashIdx: -1,
	}

	m.ht[0] = newHashtable[K, V]()
	m.ht[1] = newHashtable[K, V]()

	return m
}

// Set to set key and value in hashtable, rehashing flag will affect to
// which hashtable would save the incoming item.
func (d *Map[K, V]) Set(key K, value V) {
	if d.ht[0].table == nil {
		// true: if ht[0].table is empty, then initialize the table.
		d.ht[0].init(_INIT_TABLE_CAP)
	}

	if !d.isRehashing() && d.needRehash() {
		// true: no in rehashing status and the Dict need to rehash
		d.startrehash(_GROWUP)
	}

	if d.isRehashing() {
		// true: rehashing, continue steprehash
		d.steprehash()
	}

	hashkey := d.hashkey(key)
	if d.isRehashing() {
		// true: still in rehashing
		// put item ht[1] rather than ht[0]
		d.ht[1].insert(hashkey, newEntry[K, V](key, value, (*entry[K, V])(nil)))
		return
	}

	// not in rehashing, so just put into ht[0]
	d.ht[0].insert(hashkey, newEntry[K, V](key, value, (*entry[K, V])(nil)))
	return
}

// Lookup to get an item from hashtable
func (d *Map[K, V]) Lookup(key K) (v V, ok bool) {
	if d.isRehashing() {
		// if still in rehashing, do steprehash
		d.steprehash()
	}

	hashkey := d.hashkey(key)
	v, ok = d.ht[0].lookup(hashkey, key)
	if !d.isRehashing() {
		// true: not in rehashing, just return the value
		return
	} else if ok {
		// true: in rehashing but has got in d.ht[0]
		return v, ok
	}

	// check in ht[1]
	v, ok = d.ht[1].lookup(hashkey, key)
	return
}

// Remove to delete an item in hashtable.
func (d *Map[K, V]) Remove(key K) {
	if d.ht[0].used == 0 && d.ht[1].used == 0 {
		return
	}

	if !d.isRehashing() && d.needShrink() {
		// true: no in rehashing status and the Dict need to shrink
		d.startrehash(_SHRINK)
	}

	if d.isRehashing() {
		// if still in rehashing, do steprehash
		d.steprehash()
	}

	hashkey := d.hashkey(key)
	d.ht[0].delete(hashkey, key)

	if d.isRehashing() {
		d.ht[1].delete(hashkey, key)
	}
}

func (d *Map[K, V]) Range(fn func(key K, value V) bool) {
	d.ht[0].iter(fn)

	if d.isRehashing() {
		// true: if in rehashing, so iterate the ht[1] too.
		d.ht[1].iter(fn)
	}
}

// Len get total used size of Dict
func (d *Map[K, V]) Len() int {
	return d.ht[0].used + d.ht[1].used
}

// startrehash to apply the size for ht[1] and set rehashIdx into 0
// which means starting rehashing. The size of ht[1] should be decided
// by flag [_rehashEnum]. if flag is _GROWUP, ht[1].size should be bigger than ht[0].used, otherwise,
// size should be less than ht[0].used
func (d *Map[K, V]) startrehash(flag _rehashEnum) {
	var (
		size = d.ht[0].size
	)
	switch flag {
	case _GROWUP:
		for size <= d.ht[0].used {
			size *= 2
		}
	case _SHRINK:
		// true: find the number is bigger than used but less than size
		for size >= d.ht[0].used {
			size /= 2
		}
		size = 2 * size
	}

	// init ht[1] with size
	d.ht[1].init(size)

	// modify the flag of rehash into starting
	d.rehashIdx = 0
	// println("start rehash", d.ht[1].size, d.ht[1].used)
}

// steprehash find the first list-head which is not nil to move.
func (d *Map[K, V]) steprehash() {
	ent := d.ht[0].table[d.rehashIdx]
	for ent == nil {
		// true: current side-list is nil, find the overflow
		d.rehashIdx++
		// overflow of rehashIdx checking
		if d.rehashIdx > d.ht[0].sizemask {
			// true: no more item to move
			d.finishrehash()
			return
		}
		// move to overflow side-list
		ent = d.ht[0].table[d.rehashIdx]
	}

	next := ent.overflow
	for ent != nil {
		ent.overflow = nil
		d.ht[1].insert(d.hashkey(ent.key), ent)
		if next == nil {
			// true: the side-list has loop over
			break
		}
		// update the iterator
		ent = next
		next = next.overflow
	}

	// release the memory of ht[0].table[d.rehashIdx]
	d.ht[0].table[d.rehashIdx] = nil
	d.rehashIdx++

	if d.rehashIdx > d.ht[0].sizemask {
		// true: all items have been moved
		d.finishrehash()
	}
}

// finishrehash to finish the rehashing work with following steps:
// release ht[0].
// move ht[1] to ht[0].
// create a new hashtable and assign to ht[1].
func (d *Map[K, V]) finishrehash() {
	d.ht[0] = newHashtable[K, V]()
	d.ht[0], d.ht[1] = d.ht[1], d.ht[0]
	d.rehashIdx = -1
	// println("finish rehash ht[0]", d.ht[0].size, d.ht[0].used)
	// println("finish rehash ht[1]", d.ht[1].size, d.ht[1].used)
}

func (d *Map[K, V]) hashkey(key K) uint64 {
	x := (*[2]uintptr)(unsafe.Pointer(&key))
	b := [3]uintptr{x[0], x[1], x[1]}
	return murmur3.Sum64(*(*[]byte)(unsafe.Pointer(&b)))
}

func (d *Map[K, V]) needRehash() bool {
	// fill ration is bigger than _REHASH_GROW_THRESHOLD (100%)
	return float32(d.ht[0].used)/float32(d.ht[0].size) > _REHASH_GROW_THRESHOLD
}

func (d *Map[K, V]) needShrink() bool {
	// hashtable size is bigger than _INIT_TABLE_CAP and fill ratio is less than 10%
	return d.ht[0].size > _INIT_TABLE_CAP && (d.ht[0].used*100/d.ht[0].size) < _REHASH_SHRINK_THRESHOLD
}

func (d *Map[K, V]) isRehashing() bool {
	return d.rehashIdx != -1
}
