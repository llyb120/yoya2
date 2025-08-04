package y

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// Map 是一个协程安全的有序映射，按插入顺序维护键值对
type Map[K comparable, V any] struct {
	mu      sync.RWMutex
	keys    []K
	vals    []V // 仅在rmap下有用
	mp      map[K]V
	options struct {
		isRMap bool
	}
}

// NeworderedMap 创建一个新的有序映射
func NewMap[K comparable, V any](args ...any) *Map[K, V] {
	om := &Map[K, V]{
		mp: make(map[K]V),
	}
	for _, arg := range args {
		switch v := arg.(type) {
		case map[K]V:
			for k, v := range v {
				om.set(k, v)
			}
		case *Map[K, V]:
			v.ForEach(func(key K, value V) bool {
				om.set(key, value)
				return true
			})
		default:
			if v == RMap {
				om.options.isRMap = true
			}
		}
	}
	return om
}

// Set 添加或更新键值对
func (om *Map[K, V]) Set(key K, value V) {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.set(key, value)
}

// Get 获取键对应的值
func (om *Map[K, V]) Get(key K) (V, bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return om.get(key)
}

// RGet 获取值对应的键（仅在RMap模式下有效）
func (om *Map[K, V]) RGet(value V) (K, bool) {
	if !om.options.isRMap {
		var zero K
		return zero, false
	}

	om.mu.RLock()
	defer om.mu.RUnlock()

	// 使用vals切片进行快速查找
	for i, v := range om.vals {
		if any(v) == any(value) {
			return om.keys[i], true
		}
	}
	var zero K
	return zero, false
}

// RSet 设置键值对，并确保值的唯一性（仅在RMap模式下有效）
// 注意：在RMap中，应该使用 RSet(值, 键) 的方式调用
func (om *Map[K, V]) RSet(value V, key K) bool {
	if !om.options.isRMap {
		return false
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	// 检查值是否已存在
	for _, v := range om.vals {
		if any(v) == any(value) {
			// 值已存在，不允许重复
			return false
		}
	}

	om.set(key, value)
	return true
}

// RDel 删除指定值对应的键值对（仅在RMap模式下有效）
func (om *Map[K, V]) RDel(value V) (K, bool) {
	if !om.options.isRMap {
		var zero K
		return zero, false
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	// 使用vals切片进行查找
	for i, v := range om.vals {
		if any(v) == any(value) {
			// 找到对应的键
			key := om.keys[i]
			// 从keys和vals中删除
			om.keys = append(om.keys[:i], om.keys[i+1:]...)
			om.vals = append(om.vals[:i], om.vals[i+1:]...)
			// 从map中删除
			delete(om.mp, key)
			return key, true
		}
	}

	var zero K
	return zero, false
}

// Del 删除键值对
func (om *Map[K, V]) Del(key K) V {
	om.mu.Lock()
	defer om.mu.Unlock()

	var index = -1
	for i, k := range om.keys {
		if k == key {
			index = i
			break
		}
	}
	if index == -1 {
		var zero V
		return zero
	} else {
		val := om.mp[key]
		delete(om.mp, key)
		om.keys = append(om.keys[:index], om.keys[index+1:]...)
		return val
	}
}

// Size 返回映射大小
func (om *Map[K, V]) Size() int {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return len(om.keys)
}

// Keys 按插入顺序返回所有键
func (om *Map[K, V]) Keys() []K {
	om.mu.RLock()
	defer om.mu.RUnlock()

	keys := make([]K, len(om.keys))
	copy(keys, om.keys)
	return keys
}

// Vals 按插入顺序返回所有值
func (om *Map[K, V]) Vals() []V {
	om.mu.RLock()
	defer om.mu.RUnlock()

	values := make([]V, len(om.keys))
	for i, key := range om.keys {
		values[i] = om.mp[key]
	}
	return values
}

// Clear 清空映射
func (om *Map[K, V]) Clear() {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.clear()
}

// ForEach 按顺序遍历所有键值对
func (om *Map[K, V]) ForEach(fn func(key K, value V) bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	om.foreach(fn)
}

func (om *Map[K, V]) SortByKey(fn func(a, b K) bool) {
	om.mu.Lock()
	defer om.mu.Unlock()

	sort.Slice(om.keys, func(i, j int) bool {
		return fn(om.keys[i], om.keys[j])
	})
}

func (om *Map[K, V]) SortByValue(fn func(a, b V) bool) {
	om.mu.Lock()
	defer om.mu.Unlock()

	type pair struct {
		Key   K
		Value V
	}
	values := make([]pair, len(om.keys))
	for i, key := range om.keys {
		values[i] = pair{Key: key, Value: om.mp[key]}
	}
	sort.Slice(values, func(i, j int) bool {
		return fn(values[i].Value, values[j].Value)
	})
	keys := make([]K, len(values))
	for i, v := range values {
		keys[i] = v.Key
	}
	om.keys = keys
}

func (om *Map[K, V]) Fork() *Map[K, V] {
	om.mu.RLock()
	defer om.mu.RUnlock()

	forkMap := NewMap[K, V](om)
	return forkMap
}

func (om *Map[K, V]) Pos(key K) int {
	om.mu.RLock()
	defer om.mu.RUnlock()

	for i, k := range om.keys {
		if k == key {
			return i
		}
	}
	return -1
}

func (om *Map[K, V]) clear() {
	om.keys = nil
	om.mp = make(map[K]V)
	om.vals = nil
}

func (om *Map[K, V]) set(key K, value V) {
	if om.options.isRMap {
		// 在RMap模式下，保持vals和keys一一对应
		if index := om.Pos(key); index != -1 {
			// 键已存在，更新值
			om.vals[index] = value
		} else {
			// 新键，添加到末尾
			om.keys = append(om.keys, key)
			om.vals = append(om.vals, value)
		}
	} else if _, exists := om.mp[key]; !exists {
		om.keys = append(om.keys, key)
	}
	om.mp[key] = value
}

func (om *Map[K, V]) get(key K) (V, bool) {
	v, ok := om.mp[key]
	return v, ok
}

func (om *Map[K, V]) foreach(fn func(key K, value V) bool) {
	for _, key := range om.keys {
		if !fn(key, om.mp[key]) {
			break
		}
	}
}

func (om *Map[K, V]) lock() {
	om.mu.Lock()
}

func (om *Map[K, V]) unlock() {
	om.mu.Unlock()
}

func (om *Map[K, V]) MarshalJSON() ([]byte, error) {
	return marshalMap[K, V](om)
}

func (om *Map[K, V]) UnmarshalJSON(data []byte) error {
	return unmarshalMap[K, V](om, data)
}

type jsonMap[K any, V any] interface {
	lock()
	unlock()
	set(key K, value V)
	clear()
	foreach(fn func(key K, value V) bool)
}

func marshalMap[K any, V any](mp jsonMap[K, V]) ([]byte, error) {
	mp.lock()
	defer mp.unlock()

	var buf bytes.Buffer
	buf.WriteByte('{')

	var reultErr error
	first := true
	mp.foreach(func(key K, value V) bool {
		if !first {
			buf.WriteByte(',')
		}
		first = false

		// 序列化键
		keyBytes, err := json.Marshal(key)
		if err != nil {
			reultErr = err
			return false
		}
		buf.Write(keyBytes)

		buf.WriteByte(':')

		// 序列化值
		valueBytes, err := json.Marshal(value)
		if err != nil {
			reultErr = err
			return false
		}
		buf.Write(valueBytes)
		return true
	})

	buf.WriteByte('}')
	return buf.Bytes(), reultErr
}

// UnmarshalJSON 实现json.Unmarshaler接口
func unmarshalMap[K any, V any](mp jsonMap[K, V], data []byte) error {
	mp.lock()
	defer mp.unlock()

	// 清空现有数据
	mp.clear()

	dec := json.NewDecoder(bytes.NewReader(data))

	// 确保开始是一个对象
	if t, err := dec.Token(); err != nil {
		return err
	} else if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected {, got %v", t)
	}

	// 读取键值对
	for dec.More() {
		// 读取键
		var key K
		keyToken, err := dec.Token()
		if err != nil {
			return err
		}

		// 如果键是字符串类型，需要特殊处理
		if keyStr, ok := keyToken.(string); ok {
			if err := json.Unmarshal([]byte(`"`+keyStr+`"`), &key); err != nil {
				return err
			}
		} else {
			if err := json.Unmarshal([]byte(fmt.Sprintf("%v", keyToken)), &key); err != nil {
				return err
			}
		}

		// 读取值
		var value V
		if err := dec.Decode(&value); err != nil {
			return err
		}

		// 添加到有序映射
		mp.set(key, value)
	}

	// 确保结束是一个对象
	if t, err := dec.Token(); err != nil {
		return err
	} else if delim, ok := t.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expected }, got %v", t)
	}

	return nil
}
