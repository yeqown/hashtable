// Package hashtable to implements Map interface
// refer to https://redisbook.readthedocs.io/en/latest/internal-datastruct/dict.html
// @author yeqown
// @date 2019-11-28
//
package hashtable

import (
	"github.com/spaolacci/murmur3"
)

var (
	_ Map = &LinkedDict{}
)

type rehashTyp uint8

const (
	// rehashGrowup .
	rehashGrowup rehashTyp = iota + 1
	// rehashShrink .
	rehashShrink
)

const (
	// InitTableSize .
	InitTableSize = 4

	// RehashRatio .
	RehashRatio = 1.0

	// RehashShrinkRatio .
	RehashShrinkRatio = 10
)

// LinkedDict an Dict based linked-list
type LinkedDict struct {
	// two hashtable to work with, ht[0] will be use normally,
	// ht[1] will only be used when rehashing
	ht [2]*hashtable

	// rehashing flag, -1 means not in rehashing
	rehashIdx int
}

// NewLinkedDict to create an Dict which is implemented based on linked-list
func NewLinkedDict() Map {
	m := &LinkedDict{
		ht:        [2]*hashtable{},
		rehashIdx: -1,
	}

	m.ht[0] = newHashtable()
	m.ht[1] = newHashtable()

	return m
}

// Set to set key and value in hashtable, rehasing flag will affet to
// which hashtable would be save the incoming item.
func (d *LinkedDict) Set(key string, value interface{}) {
	if d.ht[0].table == nil {
		// true: if ht[0].table is empty, then initialize the table.
		d.ht[0].init(InitTableSize)
	}

	if !d.isRehashing() && d.needRehash() {
		// true: no in rehashing status and the Dict need to rehash
		d.startrehash(rehashGrowup)
	}

	if d.isRehashing() {
		// true: rehashing, continue steprehash
		d.steprehash()
	}

	hashkey := d.hashkey(key)
	if d.isRehashing() {
		// true: still in rehashing
		// put item ht[1] rather than ht[0]
		d.ht[1].insert(hashkey, newDictEntry(key, value, nil))
		return
	}

	// not in rehashing, so just put into ht[0]
	d.ht[0].insert(hashkey, newDictEntry(key, value, nil))
	return
}

// Get to get an item from hashtable
func (d *LinkedDict) Get(key string) (v interface{}, ok bool) {
	if d.isRehashing() {
		// if still in rehashing, do steprehash
		d.steprehash()
	}

	hashkey := d.hashkey(key)
	v, ok = d.ht[0].lookup(hashkey, key)
	if !d.isRehashing() {
		// true: not in rehashing, just return the value
		return
	}

	// check in ht[1]
	v, ok = d.ht[1].lookup(hashkey, key)
	return
}

// Del to delete an item in hashtable.
func (d *LinkedDict) Del(key string) {
	if d.ht[0].used == 0 && d.ht[1].used == 0 {
		return
	}

	if !d.isRehashing() && d.needShrink() {
		// true: no in rehashing status and the Dict need to shrink
		d.startrehash(rehashShrink)
	}

	if d.isRehashing() {
		// if still in rehashing, do steprehash
		d.steprehash()
	}

	hashkey := d.hashkey(key)
	d.ht[0].del(hashkey, key)

	if d.isRehashing() {
		d.ht[1].del(hashkey, key)
	}
}

// Iter .
func (d *LinkedDict) Iter(iterFunc MapIterator) {
	d.ht[0].iter(iterFunc)

	if d.isRehashing() {
		// true: if in rehashing, so iterate the ht[1] too.
		d.ht[1].iter(iterFunc)
	}
}

// Len get totoal used size of Dict
func (d *LinkedDict) Len() int {
	return d.ht[0].used + d.ht[1].used
}

// startrehash to apply the size for ht[1] and set rehashIdx into 0
// which means starting rehashing. The size of ht[1] should be decided
// by flag [rehashTyp]. if flag is rehashGrowup, ht[1].size should be bigger than ht[0].used, otherwise,
// size should be less than ht[0].used
func (d *LinkedDict) startrehash(flag rehashTyp) {
	var (
		size = d.ht[0].size
	)
	switch flag {
	case rehashGrowup:
		for size <= d.ht[0].used {
			size *= 2
		}
	case rehashShrink:
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

// steprehash .
// find the first list-head which is not nil to move.
func (d *LinkedDict) steprehash() {
	entry := d.ht[0].table[d.rehashIdx]
	for entry == nil {
		// true: current side-list is nil, find the next
		d.rehashIdx++
		// overflow of rehashIdx checking
		if d.rehashIdx > d.ht[0].sizemask {
			// true: no more item to move
			d.finishrehash()
			return
		}
		// move to next side-list
		entry = d.ht[0].table[d.rehashIdx]
	}

	next := entry.next
	for entry != nil {
		entry.next = nil
		d.ht[1].insert(d.hashkey(entry.key), entry)
		if next == nil {
			// true: the side-list has loop over
			break
		}
		// update the iterator
		entry = next
		next = next.next
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
// create an new hashtable and assign to ht[1].
func (d *LinkedDict) finishrehash() {
	d.ht[0] = newHashtable()
	d.ht[0], d.ht[1] = d.ht[1], d.ht[0]
	d.rehashIdx = -1
	// println("finish rehash ht[0]", d.ht[0].size, d.ht[0].used)
	// println("finish rehash ht[1]", d.ht[1].size, d.ht[1].used)
}

func (d *LinkedDict) hashkey(key string) uint64 {
	return murmur3.Sum64([]byte(key))
}

func (d *LinkedDict) needRehash() bool {
	// fill ration is bigger than RehashRatio (100%)
	return float32(d.ht[0].used)/float32(d.ht[0].size) > RehashRatio
}

func (d *LinkedDict) needShrink() bool {
	// hashtable's size is bigger than InitTableSize and fill ratio is less than 10%
	return d.ht[0].size > InitTableSize && (d.ht[0].used*100/d.ht[0].size) < RehashShrinkRatio
}

func (d *LinkedDict) isRehashing() bool {
	return (d.rehashIdx != -1)
}

type hashtable struct {
	// bucket to contains items
	table []*dictEntry

	// size of hashtable
	size int

	// mask to length of table array, usually be used to calculate index
	sizemask int

	// to mark how many items has been saved in current hashtable
	used int
}

func (ht *hashtable) init(size int) {
	ht.table = make([]*dictEntry, size)
	ht.size = size
	ht.sizemask = size - 1
}

func (ht *hashtable) insert(hashkey uint64, item *dictEntry) {
	// println(ht.size, ht.used)
	pos := hashkey % uint64(ht.size)
	// if ht.table[pos] == nil {
	// 	ht.table[pos] = item
	// 	return
	// }
	entry := ht.table[pos]
	last := entry
	if entry == nil {
		ht.used++
		ht.table[pos] = item
		return
	}

	for entry != nil {
		if entry.key == item.key {
			// true: update value
			entry.value = item.value
			return
		}
		last = entry
		entry = entry.next
	}
	ht.used++
	last.next = item
}

func (ht *hashtable) del(hashkey uint64, key string) {
	pos := hashkey % uint64(ht.size)

	entry := ht.table[pos]
	var preEntry *dictEntry

	for entry != nil {
		if entry.key == key {
			// true: hit the key
			if preEntry != nil {
				preEntry.next = entry.next
			} else {
				// ture: ht.table[pos] the first item is target
				ht.table[pos] = entry.next
			}
			ht.used--
			return
		}

		preEntry = entry
		entry = entry.next
	}
}

func (ht *hashtable) lookup(hashkey uint64, key string) (v interface{}, ok bool) {
	pos := hashkey % uint64(ht.size)
	entry := ht.table[pos]
	for entry != nil {
		if entry.key == key {
			v = entry.value
			ok = true
			return
		}
		entry = entry.next
	}
	return
}

func (ht *hashtable) iter(iter MapIterator) {
	for i := 0; i < ht.size; i++ {
		entry := ht.table[i]
		for entry != nil {
			iter(entry.key, entry.value)
			entry = entry.next
		}
	}
}

// alloc memory for hash table with some init state
func newHashtable() *hashtable {
	ht := &hashtable{
		table:    nil,
		size:     0,
		sizemask: 0,
		used:     0,
	}

	return ht
}

// item of hashtable
type dictEntry struct {
	// key of entry
	key string

	// value of entry
	value interface{}

	// pointer to next item
	next *dictEntry
}

func newDictEntry(key string, value interface{}, next *dictEntry) *dictEntry {
	return &dictEntry{
		key:   key,
		value: value,
		next:  next,
	}
}
