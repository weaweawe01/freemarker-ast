package diff

import (
	"fmt"
	"sort"
)

// Difference describes one mismatch between oracle and actual payloads.
type Difference struct {
	Path   string
	Oracle any
	Actual any
}

// JSONLike computes structural differences between two JSON-like values.
func JSONLike(oracle any, actual any) []Difference {
	var diffs []Difference
	diffValue("$", oracle, actual, &diffs)
	return diffs
}

func diffValue(path string, o any, a any, diffs *[]Difference) {
	switch ov := o.(type) {
	case map[string]any:
		av, ok := a.(map[string]any)
		if !ok {
			*diffs = append(*diffs, Difference{Path: path, Oracle: o, Actual: a})
			return
		}
		diffMap(path, ov, av, diffs)
	case []any:
		av, ok := a.([]any)
		if !ok {
			*diffs = append(*diffs, Difference{Path: path, Oracle: o, Actual: a})
			return
		}
		diffSlice(path, ov, av, diffs)
	default:
		if !equalScalar(o, a) {
			*diffs = append(*diffs, Difference{Path: path, Oracle: o, Actual: a})
		}
	}
}

func diffMap(path string, o map[string]any, a map[string]any, diffs *[]Difference) {
	keySet := map[string]struct{}{}
	for k := range o {
		keySet[k] = struct{}{}
	}
	for k := range a {
		keySet[k] = struct{}{}
	}

	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		ov, okO := o[k]
		av, okA := a[k]
		switch {
		case !okO:
			*diffs = append(*diffs, Difference{Path: fmt.Sprintf("%s.%s", path, k), Oracle: nil, Actual: av})
		case !okA:
			*diffs = append(*diffs, Difference{Path: fmt.Sprintf("%s.%s", path, k), Oracle: ov, Actual: nil})
		default:
			diffValue(fmt.Sprintf("%s.%s", path, k), ov, av, diffs)
		}
	}
}

func diffSlice(path string, o []any, a []any, diffs *[]Difference) {
	maxLen := len(o)
	if len(a) > maxLen {
		maxLen = len(a)
	}
	for i := 0; i < maxLen; i++ {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		switch {
		case i >= len(o):
			*diffs = append(*diffs, Difference{Path: elemPath, Oracle: nil, Actual: a[i]})
		case i >= len(a):
			*diffs = append(*diffs, Difference{Path: elemPath, Oracle: o[i], Actual: nil})
		default:
			diffValue(elemPath, o[i], a[i], diffs)
		}
	}
}

func equalScalar(o any, a any) bool {
	return fmt.Sprintf("%v", o) == fmt.Sprintf("%v", a)
}
