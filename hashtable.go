// Package hashtable to implements Map interface
// refer to https://redisbook.readthedocs.io/en/latest/internal-datastruct/dict.html
// @author yeqown
// @date 2019-11-28
//
package hashtable

// entry(item) of hashtable
type entry[K comparable, V any] struct {
	key      K            // key of entry
	value    V            // value of entry
	overflow *entry[K, V] // overflow pointer
}

func newEntry[K comparable, V any](key K, value V, overflow *entry[K, V]) *entry[K, V] {
	return &entry[K, V]{
		key:      key,
		value:    value,
		overflow: overflow,
	}
}

type hashtable[K comparable, V any] struct {
	// bucket contains items
	table []*entry[K, V]

	// size of hashtable
	size int

	// sizemask equals to size-1 which hashtable is initialized,
	// usually be used to calculate index.
	sizemask int

	// to mark how many items has been saved in current hashtable
	used int
}

// newHashtable alloc memory for hash table with some init state
func newHashtable[K comparable, V any]() *hashtable[K, V] {
	ht := &hashtable[K, V]{
		table:    nil,
		size:     0,
		sizemask: 0,
		used:     0,
	}

	return ht
}

func (h *hashtable[K, V]) init(size int) {
	h.table = make([]*entry[K, V], size)
	h.size = size
	h.sizemask = size - 1
}

func (h *hashtable[K, V]) keycmp(key1, key2 K) bool {
	return key1 == key2
}

func (h *hashtable[K, V]) insert(hashkey uint64, item *entry[K, V]) {
	// println(h.size, h.used)
	pos := hashkey % uint64(h.size)
	// if h.table[pos] == nil {
	// 	h.table[pos] = item
	// 	return
	// }
	entry := h.table[pos]
	last := entry
	if entry == nil {
		h.used++
		h.table[pos] = item
		return
	}

	for entry != nil {
		if h.keycmp(entry.key, item.key) {
			// true: update value
			entry.value = item.value
			return
		}
		last = entry
		entry = entry.overflow
	}
	h.used++
	last.overflow = item
}

func (h *hashtable[K, V]) delete(hashkey uint64, key K) {
	pos := hashkey % uint64(h.size)

	ent := h.table[pos]
	var pre *entry[K, V]

	for ent != nil {
		if h.keycmp(ent.key, key) {
			// true: hit the key
			if pre != nil {
				pre.overflow = ent.overflow
			} else {
				// ture: h.table[pos] the first item is the target.
				h.table[pos] = ent.overflow
			}
			h.used--
			return
		}

		pre = ent
		ent = ent.overflow
	}
}

func (h *hashtable[K, V]) lookup(hashkey uint64, key K) (v V, ok bool) {
	pos := hashkey % uint64(h.size)
	ent := h.table[pos]
	for ent != nil {
		if h.keycmp(ent.key, key) {
			v = ent.value
			ok = true
			return
		}
		ent = ent.overflow
	}
	return
}

func (h *hashtable[K, V]) iter(fn func(K, V) bool) {
	for i := 0; i < h.size; i++ {
		ent := h.table[i]
		for ent != nil {
			fn(ent.key, ent.value)
			ent = ent.overflow
		}
	}
}
