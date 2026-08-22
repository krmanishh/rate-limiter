package redislimiter

import "fmt"

// ParseInts converts the result of a Lua script that returns a table of
// n integers into a []int64, failing if the shape doesn't match.
func ParseInts(result interface{}, n int) ([]int64, error) {
	values, ok := result.([]interface{})

	if !ok || len(values) != n {
		return nil, fmt.Errorf("unexpected redis result: %v", result)
	}

	ints := make([]int64, n)

	for i, v := range values {
		asInt, ok := v.(int64)

		if !ok {
			return nil, fmt.Errorf(
				"unexpected redis result element type: %T",
				v,
			)
		}

		ints[i] = asInt
	}

	return ints, nil
}

// CeilSecondsFromMillis converts a millisecond duration (as returned by
// the Lua scripts) to whole seconds, rounding up. Truncating a
// sub-second wait down to 0 would tell a client to retry before enough
// time has actually passed.
func CeilSecondsFromMillis(ms int64) int {
	if ms <= 0 {
		return 0
	}

	return int((ms + 999) / 1000)
}
