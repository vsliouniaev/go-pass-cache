package util

import (
	"sort"
)

func SortedKeys(in map[string]struct{}) []string {
	var sorted = make([]string, len(in))
	i := 0
	for k := range in {
		sorted[i] = k
		i++
	}
	sort.Strings(sorted)
	return sorted
}
