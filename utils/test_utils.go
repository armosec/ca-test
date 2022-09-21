package utils

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func SaveExpected(t *testing.T, fileName string, i interface{}) {
	data, _ := json.MarshalIndent(i, "", "    ")
	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		panic(err)
	}
	t.Log("Updating expected file: "+fileName, " with actual response: ", string(data))
	assert.False(t, true, "update expected is true, set to false and rerun test")
}

func CompareOrUpdate[T any](actual T, expectedBytes []byte, expectedFileName string, t *testing.T, update bool) {
	if update {
		SaveExpected(t, expectedFileName, actual)
	} else {
		Equal(t, actual, expectedBytes)
	}
}

func Equal[T any](t *testing.T, actual T, expectedBytes []byte) {
	var expected T
	err := json.Unmarshal(expectedBytes, &expected)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func DeepEqualOrUpdate(t *testing.T, actualBytes, expectedBytes []byte, expectedFileName string, update bool) {
	if update {
		var actual interface{}
		err := json.Unmarshal(actualBytes, &actual)
		assert.NoError(t, err)
		SaveExpected(t, expectedFileName, actual)
	} else {
		DeepEqual(t, actualBytes, expectedBytes)
	}
}

func DeepEqual(t *testing.T, actualBytes, expectedBytes []byte) {
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
	assert.Empty(t, diff, "actual compare with expected should not have diffs")
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
