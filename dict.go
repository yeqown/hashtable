package hashtable

// Map .
type Map interface {
	// get value from Dict with key, if the key not exists it will return nil and false
	Get(key string) (v interface{}, ok bool)

	// set key and value into Dict, if has key existed, then replace the data
	Set(key string, v interface{})

	// delete a key from Dict
	Del(key string)

	// Iter func to visit all data in
	Iter(iter MapIterator)
}

// MapIterator .
type MapIterator func(key string, value interface{})
