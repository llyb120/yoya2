package y

import (
	"container/list" // 引入 container/list 包
	"sync"
	"sync/atomic"
	"time"
	"strconv"
	"strings"
)

// lruEntry represents a node in the LRU linked list.
type lruEntry[K comparable, V any] struct {
	key        K
	value      V
	expireTime *time.Time // 过期时间
	size       uint64    // 条目占用的内存大小
}

const (
	// 默认清理比例，当内存达到限制时，清理20%的过期或最旧项目
	defaultCleanupRatio = 0.2
	// 最小清理数量
	minCleanupCount = 1
)

type BaseCache[K comparable, V any] struct {
	mu            sync.RWMutex
	cache         map[K]*list.Element // Map key to list element for O(1) access
	ll            *list.List          // Doubly linked list for LRU order
	maxSize       int                 // Max number of items in the cache
	maxMemory     uint64              // 最大内存限制（字节）
	currentMemory uint64              // 当前使用的内存（原子操作）
	defaultTTL    time.Duration       // 默认过期时间，0表示永不过期
	cleanupRatio  float64             // 清理比例
}

type CacheOption struct {
	MaxSize   int           // 最大条目数，0表示不限制
	MaxMemory string        // 最大内存限制，支持 "10m", "1g" 等格式
	TTL       time.Duration // 默认过期时间，0表示永不过期
}

// parseMemory 解析内存大小字符串，如 "10m", "1g" 等
func parseMemory(sizeStr string) (uint64, error) {
	if sizeStr == "" {
		return 0, nil
	}

	sizeStr = strings.ToLower(sizeStr)
	var multiplier uint64 = 1

	switch {
	case strings.HasSuffix(sizeStr, "k"):
		multiplier = 1 << 10
		sizeStr = sizeStr[:len(sizeStr)-1]
	case strings.HasSuffix(sizeStr, "m"):
		multiplier = 1 << 20
		sizeStr = sizeStr[:len(sizeStr)-1]
	case strings.HasSuffix(sizeStr, "g"):
		multiplier = 1 << 30
		sizeStr = sizeStr[:len(sizeStr)-1]
	case strings.HasSuffix(sizeStr, "t"):
		multiplier = 1 << 40
		sizeStr = sizeStr[:len(sizeStr)-1]
	}

	size, err := strconv.ParseUint(sizeStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return size * multiplier, nil
}

func NewBaseCache[K comparable, V any](opts CacheOption) *BaseCache[K, V] {
	var maxMemory uint64
	if opts.MaxMemory != "" {
		size, err := parseMemory(opts.MaxMemory)
		if err != nil {
			panic("invalid max memory format: " + err.Error())
		}
		maxMemory = size
	}

	c := &BaseCache[K, V]{
		cache:        make(map[K]*list.Element),
		ll:           list.New(),
		maxSize:      opts.MaxSize,
		maxMemory:    maxMemory,
		defaultTTL:   opts.TTL,
		cleanupRatio: defaultCleanupRatio,
	}

	return c
}

// Set 设置缓存项，可以指定可选的TTL
// 使用示例:
//   cache.Set(key, value)                // 使用默认TTL
//   cache.Set(key, value, time.Hour)     // 指定TTL为1小时
func (c *BaseCache[K, V]) Set(key K, value V, ttl ...time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var ttlDuration time.Duration
	if len(ttl) > 0 {
		ttlDuration = ttl[0] // 只使用第一个TTL值
	} else {
		ttlDuration = c.defaultTTL
	}

	c.setWithTTL(key, value, ttlDuration)
}

// calculateSize 计算值的大小（字节数）
func calculateSize(v interface{}) uint64 {
	// 这里可以使用更精确的方法来计算大小
	// 这里简化为固定大小加上值的类型大小
	size := uint64(0)
	switch v := v.(type) {
	case string:
		size = uint64(len(v))
	case []byte:
		size = uint64(len(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		size = 8 // 简单处理，实际大小可能不同
	default:
		// 对于复杂类型，使用固定大小
		size = 64
	}
	return size
}

// setWithTTL 内部方法，设置带有TTL的缓存项
func (c *BaseCache[K, V]) setWithTTL(key K, value V, ttl time.Duration) {
	// 计算新条目大小
	entrySize := calculateSize(key) + calculateSize(value)

	// 如果设置了内存限制，检查是否有足够空间
	if c.maxMemory > 0 && entrySize > c.maxMemory {
		// 单个条目就超过了最大内存限制
		return
	}

	// 移除已存在的条目（如果存在）
	if existingEntry, ok := c.cache[key]; ok {
		oldEntry := existingEntry.Value.(*lruEntry[K, V])
		c.ll.Remove(existingEntry)
		atomic.AddUint64(&c.currentMemory, ^(oldEntry.size - 1)) // 减去旧条目大小
		delete(c.cache, key)
	}

	// 检查是否需要清理
	if c.maxMemory > 0 && atomic.LoadUint64(&c.currentMemory)+entrySize > c.maxMemory {
		c.mu.Unlock()
		c.maybeCleanup()
		c.mu.Lock()
	}

	// 如果设置了内存限制，且仍然没有足够空间，尝试直接清理最旧的项目
	if c.maxMemory > 0 && atomic.LoadUint64(&c.currentMemory)+entrySize > c.maxMemory && c.ll.Len() > 0 {
		back := c.ll.Back()
		if back != nil {
			entry := back.Value.(*lruEntry[K, V])
			delete(c.cache, entry.key)
			c.ll.Remove(back)
			atomic.AddUint64(&c.currentMemory, ^(entry.size - 1))
		}
	}

	// 添加新条目
	var expireTime *time.Time
	if ttl > 0 {
		expire := time.Now().Add(ttl)
		expireTime = &expire
	}

	newEntry := &lruEntry[K, V]{
		key:        key,
		value:      value,
		expireTime: expireTime,
		size:       entrySize,
	}
	element := c.ll.PushFront(newEntry)
	c.cache[key] = element
	atomic.AddUint64(&c.currentMemory, entrySize)

	// 如果设置了最大条目数限制，移除最老的条目
	for c.maxSize > 0 && c.ll.Len() > c.maxSize {
		back := c.ll.Back()
		if back == nil {
			break
		}
		entry := back.Value.(*lruEntry[K, V])
		delete(c.cache, entry.key)
		c.ll.Remove(back)
		atomic.AddUint64(&c.currentMemory, ^(entry.size - 1))
	}
}

// SetMap 批量设置缓存项
func (c *BaseCache[K, V]) SetMap(m map[K]V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, value := range m {
		c.setWithTTL(key, value, c.defaultTTL)
	}
}

func (c *BaseCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.get(key)
}

func (c *BaseCache[K, V]) Gets(keys ...K) []V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	values := make([]V, 0, len(keys))
	for _, key := range keys {
		value, ok := c.get(key)
		if ok {
			values = append(values, value)
		}
	}
	return values
}

func (c *BaseCache[K, V]) get(key K) (V, bool) {
	c.mu.RLock()
	entry, ok := c.cache[key]
	c.mu.RUnlock()

	if !ok {
		var zero V
		return zero, false
	}

	item := entry.Value.(*lruEntry[K, V])

	// 检查是否过期
	now := time.Now()
	if item.expireTime != nil && now.After(*item.expireTime) {
		c.mu.Lock()
		// 再次检查，防止并发问题
		if e, ok := c.cache[key]; ok && e == entry {
			c.ll.Remove(entry)
			delete(c.cache, key)
			atomic.AddUint64(&c.currentMemory, ^(item.size - 1))
		}
		c.mu.Unlock()
		var zero V
		return zero, false
	}

	c.mu.Lock()
	c.ll.MoveToFront(entry) // 移动到前面表示最近使用
	c.mu.Unlock()

	return item.value, true
}

func (c *BaseCache[K, V]) Del(key ...K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, k := range key {
		if entry, ok := c.cache[k]; ok {
			c.ll.Remove(entry)
			delete(c.cache, k)
		}
	}
}

func (c *BaseCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[K]*list.Element)
	c.ll.Init() // Clear the list
}

// maybeCleanup 在需要时清理过期或最旧的项目
func (c *BaseCache[K, V]) maybeCleanup() {
	// 如果内存使用未达到限制，不进行清理
	if c.maxMemory == 0 || atomic.LoadUint64(&c.currentMemory) <= c.maxMemory {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0
	targetCount := int(float64(c.ll.Len()) * c.cleanupRatio)
	if targetCount < minCleanupCount {
		targetCount = minCleanupCount
	}

	// 从链表尾部开始清理（最旧的项目）
	for e := c.ll.Back(); e != nil && removed < targetCount; {
		entry := e.Value.(*lruEntry[K, V])
		next := e.Prev()

		// 如果项目已过期或需要释放内存，则删除
		if (entry.expireTime != nil && now.After(*entry.expireTime)) || 
		   atomic.LoadUint64(&c.currentMemory) > c.maxMemory {
			delete(c.cache, entry.key)
			c.ll.Remove(e)
			atomic.AddUint64(&c.currentMemory, ^(entry.size - 1))
			removed++
		} else if removed == 0 {
			// 如果第一个项目未过期，且内存仍然超限，强制清理最旧的项目
			if atomic.LoadUint64(&c.currentMemory) > c.maxMemory {
				delete(c.cache, entry.key)
				c.ll.Remove(e)
				atomic.AddUint64(&c.currentMemory, ^(entry.size - 1))
				removed++
			} else {
				// 内存已足够，退出循环
				break
			}
		}

		e = next
	}
}

// SetWithTTL 设置带有过期时间的缓存项（兼容旧代码）
// 注意：推荐使用 Set(key, value, ttl) 替代
func (c *BaseCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setWithTTL(key, value, ttl)
}

// MemoryUsage 返回当前缓存使用的内存大小（字节）
func (c *BaseCache[K, V]) MemoryUsage() uint64 {
	return atomic.LoadUint64(&c.currentMemory)
}

// MemoryLimit 返回内存限制（字节）
func (c *BaseCache[K, V]) MemoryLimit() uint64 {
	return c.maxMemory
}

// SetMemoryLimit 设置内存限制
func (c *BaseCache[K, V]) SetMemoryLimit(limit string) error {
	size, err := parseMemory(limit)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.maxMemory = size

	// 如果新限制小于当前使用量，移除最老的条目直到满足限制
	for c.maxMemory > 0 && c.currentMemory > c.maxMemory && c.ll.Len() > 0 {
		back := c.ll.Back()
		if back == nil {
			break
		}
		entry := back.Value.(*lruEntry[K, V])
		delete(c.cache, entry.key)
		c.ll.Remove(back)
		atomic.AddUint64(&c.currentMemory, ^(entry.size - 1))
	}

	return nil
}

func (c *BaseCache[K, V]) Destroy() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Clear()
}

// GetOrSetFunc 获取或设置缓存项，如果不存在则调用函数生成值
func (c *BaseCache[K, V]) GetOrSetFunc(key K, fn func() V) V {
	value, ok := c.Get(key)
	if !ok {
		c.mu.Lock()
		defer c.mu.Unlock()
		if value, ok = c.get(key); ok {
			return value
		}
		value = fn()
		c.setWithTTL(key, value, c.defaultTTL)
		return value
	}
	return value
}

// Len returns the number of items in the cache.
func (c *BaseCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ll.Len()
}

// Cap returns the maximum capacity of the cache.
func (c *BaseCache[K, V]) Cap() int {
	return c.maxSize
}
