package directory_checksum

import (
	"reflect"
	"testing"
)

func TestSortedKeys(t *testing.T) {
	var m = map[string]string{
		"A": "A",
		"C": "C",
		"B": "B",
		"E": "E",
		"F": "F",
		"D": "D",
	}

	unordered_keys := make([]string, len(m))
	i := 0
	for k := range m {
		unordered_keys[i] = k
		i++
	}

	want := []string{"A", "B", "C", "D", "E", "F"}

	if reflect.DeepEqual(want, unordered_keys) {
		t.Fatalf("Unordered keys were already ordered, this may never happen")
	}

	got := sortedKeys(m)
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Got %v, want %v", got, want)
	}
}
