// Package hashtable to implements Map interface
// refer to https://redisbook.readthedocs.io/en/latest/internal-datastruct/dict.html
// @author yeqown
// @date 2019-11-28
//
package hashtable

import (
	"github.com/spaolacci/murmur3"
)

const (
	// InitTableSize .
	InitTableSize = 4

	// RehashRatio .
	RehashRatio = 1
)

// LinkedDict .
type LinkedDict struct {
	ht        [2]*hashtable
	rehashIdx int // -1
}

// NewLinkedDict .
func NewLinkedDict() Map {
	m := &LinkedDict{
		ht:        [2]*hashtable{},
		rehashIdx: -1,
	}

	m.ht[0] = newHashtable()
	m.ht[1] = newHashtable()

	return m
}

// Set .
func (d *LinkedDict) Set(key string, value interface{}) {

	if d.ht[0].table == nil {
		// true: 第一次添加
		d.ht[0].init(InitTableSize)
	}

	if !d.isRehashing() && d.needRehash() {
		// true: 不再rehash中, 且需要rehash
		d.startrehash()
	}

	if d.isRehashing() {
		// true: rehashing, continue steprehash
		d.steprehash()
	}

	hashkey := d.hashkey(key)
	if d.isRehashing() {
		// 不再往ht[0]中增加元素, 而是放在ht[1]中去
		d.ht[1].insert(hashkey, newDictEntry(key, value, nil))
		return
	}

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
func (d *LinkedDict) Iter(iter MapIterator) {

}

func (d *LinkedDict) startrehash() {
	// 为d.ht[1]申请比d.ht[0]大的空间
	// 大小至少为 ht[0]->used 的两倍；
	var (
		size = InitTableSize
	)
	for size <= d.ht[0].used {
		size *= 2
	}
	d.ht[1].init(size)
	// 修改rehash标志为开始
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
			// true: no more
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
			// true: 侧链遍历完成
			break
		}
		// 迭代
		entry = next
		next = next.next
	}

	d.ht[0].table[d.rehashIdx] = nil
	d.rehashIdx++

	if d.rehashIdx > d.ht[0].sizemask {
		// true: 全部搬迁完成
		d.finishrehash()
	}
}

// finishrehash 完成的收尾工作
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
	return d.ht[0].used/d.ht[0].size > RehashRatio
}

func (d *LinkedDict) isRehashing() bool {
	return (d.rehashIdx != -1)
}

type hashtable struct {
	table    []*dictEntry // bucket
	size     int          // 总容量
	sizemask int          // 指针数组的长度掩码，用于计算索引值
	used     int          // 哈希表现有的节点数量
}

func (ht *hashtable) init(size int) {
	ht.table = make([]*dictEntry, size)
	ht.size = size
	ht.sizemask = size - 1
}

func (ht *hashtable) insert(hashkey uint64, item *dictEntry) {
	// println(ht.size, ht.used)
	ht.used++
	pos := hashkey % uint64(ht.size)
	if ht.table[pos] == nil {
		ht.table[pos] = item
		return
	}
	entry := ht.table[pos]
	last := entry
	for entry != nil {
		if entry.key == item.key {
			// true: 更新
			entry.value = item.value
			return
		}
		last = entry
		entry = entry.next
	}
	last.next = item
}

func (ht *hashtable) del(hashkey uint64, key string) {
	pos := hashkey % uint64(ht.size)

	entry := ht.table[pos]
	for entry != nil {
		if entry.key == key {
			// true: hit
			entry = entry.next
			return
		}

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

// ht[0].table 在第一次赋值的时候初始化
// ht[1].table 在 rehash 时分配
func newHashtable() *hashtable {
	ht := &hashtable{
		table:    nil,
		size:     0,
		sizemask: 0,
		used:     0,
	}

	return ht
}

type dictEntry struct {
	key   string
	value interface{}

	next *dictEntry
}

func newDictEntry(key string, value interface{}, next *dictEntry) *dictEntry {
	return &dictEntry{
		key:   key,
		value: value,
		next:  next,
	}
}
