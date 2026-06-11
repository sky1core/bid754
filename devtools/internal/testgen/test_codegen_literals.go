package testgen

import (
	"fmt"
	"sort"
	"strings"
)

func stringIntMapLiteral(values map[string]int) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, key := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "\t%q: %d,", key, values[key])
	}
	return b.String()
}

func intIntMapLiteral(values map[int]int) string {
	keys := make([]int, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	var b strings.Builder
	for i, key := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "\t%d: %d,", key, values[key])
	}
	return b.String()
}
