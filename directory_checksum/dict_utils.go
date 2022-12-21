package directory_checksum

import "sort"

// See https://stackoverflow.com/questions/18342784/how-to-iterate-through-a-map-in-golang-in-order
type Ordered interface {
	string
}

// sortedKeys returns an alphabetically-sorted array slice of the keys of the map m.
func sortedKeys[K Ordered, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
