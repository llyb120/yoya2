package test

import (
	"fmt"
	"testing"

	"github.com/llyb120/yoya2/y"
	"github.com/stretchr/testify/assert"
)

// TestIsFilter 测试包含过滤
func TestIsFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		filter   []int
		expected []int
	}{
		{"空切片", []int{}, []int{1, 2}, []int{}},
		{"空过滤条件", []int{1, 2, 3}, []int{}, []int{1, 2, 3}},
		{"完全匹配", []int{1, 2, 3, 4, 5}, []int{1, 3, 5}, []int{1, 3, 5}},
		{"部分匹配", []int{1, 2, 3, 4, 5}, []int{2, 4, 6}, []int{2, 4}},
		{"无匹配", []int{1, 2, 3}, []int{4, 5}, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := y.Filter(tt.input, y.Is, toAnySlice(tt.filter)...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNotFilter 测试排除过滤
func TestNotFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		filter   []int
		expected []int
	}{
		{"空切片", []int{}, []int{1, 2}, []int{}},
		{"空过滤条件", []int{1, 2, 3}, []int{}, []int{1, 2, 3}},
		{"完全排除", []int{1, 2, 3, 4, 5}, []int{1, 3, 5}, []int{2, 4}},
		{"部分排除", []int{1, 2, 3, 4, 5}, []int{2, 4, 6}, []int{1, 3, 5}},
		{"无排除", []int{1, 2, 3}, []int{4, 5}, []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := y.Filter(tt.input, y.Not, toAnySlice(tt.filter)...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCombinedFilter 测试组合过滤
func TestCombinedFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		filter   []any
		expected []int
	}{
		{
			"IS和NOT组合",
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			[]any{y.Not, 5, y.Is, 1, 2, 3},
			[]int{1, 2, 3},
		},
		{
			"多个NOT",
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			[]any{y.Not, 5, y.Not, 7, y.Not, 9},
			[]int{1, 2, 3, 4, 6, 8, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := y.Filter(tt.input, y.Not, toAnySlice(tt.filter)...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFunctionFilter 测试函数过滤
func TestFunctionFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		filter   func(int) bool
		expected []int
	}{
		{"偶数", []int{1, 2, 3, 4, 5}, func(v int) bool { return v%2 == 0 }, []int{2, 4}},
		{"大于3", []int{1, 2, 3, 4, 5}, func(v int) bool { return v > 3 }, []int{4, 5}},
		{"空结果", []int{1, 2, 3}, func(v int) bool { return v > 10 }, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := y.Filter(tt.input, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStringFilter 测试字符串切片过滤
func TestStringFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		filter   []any
		expected []string
	}{
		{
			"包含特定字符串",
			[]string{"apple", "banana", "cherry", "date"},
			[]any{y.Is, "apple", "cherry"},
			[]string{"apple", "cherry"},
		},
		{
			"排除特定字符串",
			[]string{"apple", "banana", "cherry", "date"},
			[]any{y.Not, "banana", "date"},
			[]string{"apple", "cherry"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []string
			if tt.filter[0] == y.Is {
				result = y.Filter(tt.input, y.Is, tt.filter[1:]...)
			} else {
				result = y.Filter(tt.input, y.Not, tt.filter[1:]...)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCustomTypeFilter 测试自定义类型过滤
type user struct {
	ID   int
	Name string
	Age  int
}

func TestCustomTypeFilter(t *testing.T) {
	users := []user{
		{1, "Alice", 25},
		{2, "Bob", 30},
		{3, "Charlie", 20},
		{4, "David", 35},
	}

	t.Run("过滤年龄大于25的用户", func(t *testing.T) {
		result := y.Filter(users, func(u user) bool {
			return u.Age > 25
		})
		expected := []user{{2, "Bob", 30}, {4, "David", 35}}
		assert.Equal(t, expected, result)
	})

	t.Run("过滤特定名称的用户", func(t *testing.T) {
		result := y.Filter(users, func(u user) bool {
			return u.Name == "Alice" || u.Name == "Charlie"
		})
		expected := []user{{1, "Alice", 25}, {3, "Charlie", 20}}
		assert.Equal(t, expected, result)
	})
}

// TestEdgeCases 测试边界情况
func TestEdgeCases(t *testing.T) {
	t.Run("空切片", func(t *testing.T) {
		result := y.Filter([]int{}, y.Is, 1, 2, 3)
		assert.Empty(t, result)
	})

	t.Run("nil切片", func(t *testing.T) {
		var s []int = nil
		result := y.Filter(s, y.Is, 1, 2, 3)
		assert.Empty(t, result)
	})

}

func TestFilter(t *testing.T) {
	a := []int{1, 2, 3}
	b := y.Filter(a, func(v int) bool {
		return v == 1
	})
	c := y.Filter(a, func(v int, i int) bool {
		return v == 1
	})
	d := y.Filter(a, func(v *int) bool {
		*v = 2
		return true
	})
	e := y.Filter(a, func(v *int, i int) bool {
		*v = i
		return true
	})
	f := y.Filter(a, y.Is, 1)
	fmt.Println(a, b, c, d, e, f)
}

// 辅助函数：将任意类型的切片转换为[]any
func toAnySlice[T any](s []T) []any {
	if s == nil {
		return nil
	}
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}
