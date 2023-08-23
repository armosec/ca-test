package utils

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func SaveExpected(t *testing.T, fileName string, i interface{}) {
	data, _ := json.MarshalIndent(i, "", "    ")
	err := os.WriteFile(fileName, data, 0644)
	if err != nil {
		panic(err)
	}
	t.Log("Updating expected file: "+fileName, " with actual response: ", string(data))
	assert.False(t, true, "update expected is true, set to false and rerun test")
}

func CompareAndUpdate[T any](t *testing.T, actual T, expectedBytes []byte, expectedFileName string, update bool, compareOptions ...cmp.Option) {
	expected := LoadJson[T](t, expectedBytes)
	diff := cmp.Diff(expected, actual, compareOptions...)
	assert.Empty(t, diff, "expected to have no diff")
	if update && diff != "" {
		SaveExpected(t, expectedFileName, actual)
	}
	assert.False(t, update, "update expected is true, set to false and rerun test")
}

func LoadJson[T any](t *testing.T, jsonBytes []byte) T {
	var obj T
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		t.Fatal(err)
	}
	return obj
}

// Eventually will run the function f every 500ms until it returns true or timeoutSeconds is reached
// returns true if f returned true before timeoutSeconds
func Eventually(f func() bool, timeoutSeconds int) bool {
	timeout := time.After(time.Second * time.Duration(timeoutSeconds))
	tick := time.Tick(time.Millisecond * 500)
	for {
		select {
		case <-timeout:
			return false
		case <-tick:
			if f() {
				return true
			}
		}
	}
}
