package test

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/llyb120/yoya2/y"
)

func TestMapBasic(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	result := y.Map(arr, func(item int, index int) int {
		return item * 2
	})
	expected := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestMapBasic failed, expected %v, got %v", expected, result)
	}

	strArr := []string{"a", "b", "c"}
	strResult := y.Map(strArr, func(item string, index int) string {
		return item + item
	})
	strExpected := []string{"aa", "bb", "cc"}
	if !reflect.DeepEqual(strResult, strExpected) {
		t.Errorf("TestMapBasic string failed, expected %v, got %v", strExpected, strResult)
	}
}

func TestMapAsync(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	processedCount := int32(0)
	result := y.Map(arr, func(item int, index int) int {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		atomic.AddInt32(&processedCount, 1)
		return item * 2
	}, y.UseAsync)
	expected := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestMapAsync failed, expected %v, got %v", expected, result)
	}
	if processedCount != int32(len(arr)) {
		t.Errorf("TestMapAsync failed, expected %d items processed, got %d", len(arr), processedCount)
	}
}

func TestMapDistinct(t *testing.T) {
	arr := []int{1, 2, 2, 3, 1, 4}
	result := y.Map(arr, func(item int, index int) int {
		return item * 2
	}, y.UseDistinct)
	expected := []int{2, 4, 6, 8} // Order might not be guaranteed for distinct
	if len(result) != len(expected) {
		t.Errorf("TestMapDistinct length failed, expected %d, got %d", len(expected), len(result))
	}
	// Convert to map for easy comparison of elements
	resultMap := make(map[int]bool)
	for _, v := range result {
		resultMap[v] = true
	}
	for _, v := range expected {
		if _, ok := resultMap[v]; !ok {
			t.Errorf("TestMapDistinct missing element: %d", v)
		}
	}

	strArr := []string{"a", "b", "a", "c"}
	strResult := y.Map(strArr, func(item string, index int) string {
		return item
	}, y.UseDistinct)
	strExpected := []string{"a", "b", "c"}
	if len(strResult) != len(strExpected) {
		t.Errorf("TestMapDistinct string length failed, expected %d, got %d", len(strExpected), len(strResult))
	}
	strResultMap := make(map[string]bool)
	for _, v := range strResult {
		strResultMap[v] = true
	}
	for _, v := range strExpected {
		if _, ok := strResultMap[v]; !ok {
			t.Errorf("TestMapDistinct string missing element: %s", v)
		}
	}
}

func TestMapIgnoreNil(t *testing.T) {
	arr := []*int{intPtr(1), nil, intPtr(2), nil}
	result := y.Map(arr, func(item *int, index int) *int {
		if item == nil {
			return nil
		}
		val := *item * 2
		return &val
	}, y.NotNil)
	expected := []*int{intPtr(2), intPtr(4)}
	if len(result) != len(expected) {
		t.Errorf("TestMapIgnoreNil length failed, expected %d, got %d", len(expected), len(result))
	}
	for i := range result {
		if *result[i] != *expected[i] {
			t.Errorf("TestMapIgnoreNil failed, expected %v, got %v", expected, result)
		}
	}

	interfaceArr := []*string{strPtr("hello"), nil, strPtr("world")}
	interfaceResult := y.Map(interfaceArr, func(item *string, index int) *string {
		return item
	}, y.NotNil)
	if len(interfaceResult) != 2 {
		t.Errorf("TestMapIgnoreNil interface failed, expected 2, got %d", len(interfaceResult))
	}
	if fmt.Sprintf("%v", interfaceResult[0]) != fmt.Sprintf("%v", strings.NewReader("hello")) ||
		fmt.Sprintf("%v", interfaceResult[1]) != fmt.Sprintf("%v", strings.NewReader("world")) {
		t.Errorf("TestMapIgnoreNil interface failed, result not as expected")
	}
}

func TestMapIgnoreEmpty(t *testing.T) {
	arr := []string{"hello", "", "world", " "}
	result := y.Map(arr, func(item string, index int) string {
		return item
	}, y.NotEmpty)
	expected := []string{"hello", "world", " "}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestMapIgnoreEmpty failed, expected %v, got %v", expected, result)
	}

	intArr := []int{1, 0, 2, 0, 3}
	intResult := y.Map(intArr, func(item int, index int) int {
		return item
	}, y.NotEmpty)
	intExpected := []int{1, 2, 3}
	if !reflect.DeepEqual(intResult, intExpected) {
		t.Errorf("TestMapIgnoreEmpty int failed, expected %v, got %v", intExpected, intResult)
	}

	structArr := []struct{ Val string }{{Val: "a"}, {Val: ""}}
	structResult := y.Map(structArr, func(item struct{ Val string }, index int) struct{ Val string } {
		return item
	}, y.NotEmpty)
	structExpected := []struct{ Val string }{{Val: "a"}}
	if !reflect.DeepEqual(structResult, structExpected) {
		t.Errorf("TestMapIgnoreEmpty struct failed, expected %v, got %v", structExpected, structResult)
	}
}

func TestMapCombinedOptions(t *testing.T) {
	arr := []int{1, 2, 2, 0, 3, 1, 0, 4}
	result := y.Map(arr, func(item int, index int) int {
		time.Sleep(5 * time.Millisecond)
		return item * 2
	}, y.UseAsync, y.UseDistinct, y.NotEmpty)
	expected := []int{2, 4, 6, 8} // Order might not be guaranteed for distinct
	if len(result) != len(expected) {
		t.Errorf("TestMapCombinedOptions length failed, expected %d, got %d", len(expected), len(result))
	}
	resultMap := make(map[int]bool)
	for _, v := range result {
		resultMap[v] = true
	}
	for _, v := range expected {
		if _, ok := resultMap[v]; !ok {
			t.Errorf("TestMapCombinedOptions missing element: %d", v)
		}
	}
}

func TestMapPanicRecovery(t *testing.T) {
	arr := []int{1, 2, 3}
	result := y.Map(arr, func(item int, index int) int {
		if item == 2 {
			panic("test panic")
		}
		return item * 2
	}, y.UseAsync)
	// Expect default/zero value for the panicking element
	// The exact order might vary due to async, but content should be 2, 0, 6 (or equivalent)
	expectedNonZeroCount := 2
	nonZeroCount := 0
	for _, v := range result {
		if v != 0 { // Default int value
			nonZeroCount++
		}
	}
	if nonZeroCount != expectedNonZeroCount {
		t.Errorf("TestMapPanicRecovery failed, expected %d non-zero elements, got %d", expectedNonZeroCount, nonZeroCount)
	}
	// Check if the panic was recovered for the specific index
	if result[1] != 0 { // Assuming index 1 (value 2) was the one that panicked
		t.Errorf("TestMapPanicRecovery failed to recover panic for index 1, got %v", result[1])
	}
}

func TestMapEmptyArray(t *testing.T) {
	arr := []int{}
	result := y.Map(arr, func(item int, index int) int {
		return item * 2
	})
	if len(result) != 0 {
		t.Errorf("TestMapEmptyArray failed, expected empty, got %v", result)
	}
}

func TestMapIndex(t *testing.T) {
	arr := []string{"a", "b", "c"}
	result := y.Map(arr, func(item string, index int) string {
		return fmt.Sprintf("%s-%d", item, index)
	})
	expected := []string{"a-0", "b-1", "c-2"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestMapIndex failed, expected %v, got %v", expected, result)
	}
}

func TestMapTypeConversion(t *testing.T) {
	arr := []int{1, 2, 3}
	result := y.Map(arr, func(item int, index int) string {
		return strconv.Itoa(item)
	})
	expected := []string{"1", "2", "3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestMapTypeConversion failed, expected %v, got %v", expected, result)
	}
}

// Helper for TestMapIgnoreNil
func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}
