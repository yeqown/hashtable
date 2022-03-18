package examples

import "github.com/yeqown/hashtable"

func Example() {
	var m *hashtable.Map[string, string]
	m = hashtable.New[string, string]()

	m.Set("key", "value")
	m.Lookup("key")
}
