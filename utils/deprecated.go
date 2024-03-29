package utils

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

// Deprecated: use CompareAndUpdate
func CompareOrUpdate[T any](actual T, expectedBytes []byte, expectedFileName string, t *testing.T, update bool) {
	if !Equal(t, actual, expectedBytes) && update {
		SaveExpected(t, expectedFileName, actual)
	} else if update {
		assert.False(t, true, "update expected is true, set to false and rerun test")
	}
}

// Deprecated: use only for elastic response
func Equal[T any](t *testing.T, actual T, expectedBytes []byte) bool {
	var expected T
	err := json.Unmarshal(expectedBytes, &expected)
	assert.NoError(t, err)
	return assert.Equal(t, expected, actual)
}

// Deprecated: use only for elastic response
func DeepEqualOrUpdate(t *testing.T, actualBytes, expectedBytes []byte, expectedFileName string, update bool) {
	var actual, expected interface{}
	err := json.Unmarshal(actualBytes, &actual)
	assert.NoError(t, err)
	err = json.Unmarshal(expectedBytes, &expected)
	assert.NoError(t, err)

	sortSlices := cmpopts.SortSlices(func(a, b interface{}) bool {
		if less, ok := isLess(a, b); ok {
			return less
		}
		t.Fatal("don't know how to sort these types")
		return true
	})
	diff := cmp.Diff(expected, actual, sortSlices)
	if update {
		if diff != "" {
			var actual interface{}
			err := json.Unmarshal(actualBytes, &actual)
			assert.NoError(t, err)
			SaveExpected(t, expectedFileName, actual)
		} else {
			assert.False(t, true, "update expected is true, set to false and rerun test")
		}
	} else {
		assert.Empty(t, diff, "actual compare with expected should not have diffs")
	}
}

//TODO improve this compare - works for Elastic response

func isLess(a, b interface{}) (bool, bool) {
	strA, okA := a.(string)
	strB, okB := b.(string)
	if okA && okB {
		return strA < strB, true
	}
	if less, ok := lessSlice(a, b); ok {
		return less, true
	} else if less, ok := lessMapStr2Interface(a, b); ok {
		return less, true
	}

	return false, false

}

func lessSlice(a, b interface{}) (less bool, ok bool) {
	if less, ok := lessStringSlice(a, b); ok {
		return less, true
	} else if less, ok := lessInterfaceSlice(a, b); ok {
		return less, true
	}
	return false, false
}

func lessStringSlice(a, b interface{}) (less bool, ok bool) {
	strSliceA, okA := a.([]string)
	strSliceB, okB := b.([]string)
	if !okA || !okB {
		return false, false
	}
	if len(strSliceA) != len(strSliceB) {
		return len(strSliceA) < len(strSliceB), true
	}
	for i := range strSliceA {
		if strSliceA[i] != strSliceB[i] {
			return strings.Compare(strSliceA[i], strSliceB[i]) == -1, true
		}
	}
	return false, true
}

func lessInterfaceSlice(a, b interface{}) (less bool, ok bool) {
	interSliceA, okA := a.([]interface{})
	interSliceB, okB := b.([]interface{})
	if !okA || !okB {
		return false, false
	}
	if len(interSliceA) != len(interSliceB) {
		return len(interSliceA) < len(interSliceB), true
	}
	for i := range interSliceA {
		if less, ok := isLess(interSliceA[i], interSliceB[i]); !ok {
			return false, false
		} else if less {
			return true, true
		}
	}
	return false, true
}

func lessMapStr2Interface(a, b interface{}) (bool, bool) {
	mapA, okA := a.(map[string]interface{})
	mapB, okB := b.(map[string]interface{})

	if !okA || !okB {
		return false, false
	}

	if len(mapA) != len(mapB) {
		return len(mapA) < len(mapB), true
	}

	keysA := make([]string, 0, len(mapA))
	for k := range mapA {
		keysA = append(keysA, k)
	}
	keysB := make([]string, 0, len(mapB))
	for k := range mapB {
		keysB = append(keysB, k)
	}

	sort.StringSlice(keysA).Sort()
	sort.StringSlice(keysB).Sort()

	for i := range keysA {
		if keysA[i] != keysB[i] {
			return keysA[i] < keysB[i], true
		}
	}
	for i := range keysA {
		less, ok := isLess(mapA[keysA[i]], mapB[keysA[i]])
		if !ok {
			return false, false
		}
		if less {
			return true, true
		}
	}
	return false, true
}

func loadJson[T any](jsonBytes []byte) T {
	var obj T
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		panic(err)
	}
	return obj
}
