package y

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type selector struct {
	src string
	len int
	idx int
}

type selectorNode struct {
	key   string
	props map[string]*selectorProp
}

var (
	opEqual = "=="
	opLike  = "*="
	opNot   = "!="
	opGt    = ">"
	opGe    = ">="
	opLt    = "<"
	opLe    = "<="
	opErr   = "err"
)

type selectorProp struct {
	key   string
	op    string
	value any
}

func (s *selectorNode) setProp(key string, value string) {
	if strings.HasPrefix(value, opLike) {
		s.props[key] = &selectorProp{
			key:   key,
			op:    opLike,
			value: value[2:],
		}
	} else if strings.HasPrefix(value, opNot) {
		s.props[key] = &selectorProp{
			key:   key,
			op:    opNot,
			value: value[2:],
		}
	} else if strings.HasPrefix(value, opGe) {
		val, ok := toFloat64(value[2:])
		s.props[key] = &selectorProp{
			key:   key,
			op:    opGe,
			value: val,
		}
		if !ok {
			s.props[key].op = opErr
		}
	} else if strings.HasPrefix(value, opGt) {
		val, ok := toFloat64(value[1:])
		s.props[key] = &selectorProp{
			key:   key,
			op:    opGt,
			value: val,
		}
		if !ok {
			s.props[key].op = opErr
		}
	} else if strings.HasPrefix(value, opLe) {
		val, ok := toFloat64(value[2:])
		s.props[key] = &selectorProp{
			key:   key,
			op:    opLe,
			value: val,
		}
		if !ok {
			s.props[key].op = opErr
		}
	} else if strings.HasPrefix(value, opLt) {
		val, ok := toFloat64(value[1:])
		s.props[key] = &selectorProp{
			key:   key,
			op:    opLt,
			value: val,
		}
		if !ok {
			s.props[key].op = opErr
		}
	} else {
		s.props[key] = &selectorProp{
			key:   key,
			op:    opEqual,
			value: value[1:],
		}
	}
}

func (s *selector) parse() []*selectorNode {
	s.src += " "
	s.len = len(s.src)
	s.idx = 0
	var buf bytes.Buffer
	var nodes []*selectorNode
	var current *selectorNode
	for s.idx < s.len {
		c := s.src[s.idx]
		if s.isWord(c) {
			if current == nil {
				current = &selectorNode{
					key:   "",
					props: make(map[string]*selectorProp),
				}
				nodes = append(nodes, current)
			}
			buf.WriteByte(c)
		} else if s.isSpace(c) {
			if current != nil {
				if buf.Len() > 0 {
					current.key = buf.String()
					buf.Reset()
				}
				current = nil
			}
		} else if c == '[' {
			s.idx++
			if current == nil {
				current = &selectorNode{
					key:   "",
					props: make(map[string]*selectorProp),
				}
				nodes = append(nodes, current)
			} else {
				if buf.Len() > 0 {
					current.key = buf.String()
					buf.Reset()
				}
			}
			// 读取到]
			for s.idx < s.len {
				if s.src[s.idx] != ']' {
					buf.WriteByte(s.src[s.idx])
					s.idx++
				} else {
					s.parseExpr(current, buf.String())
					buf.Reset()
					current = nil
					break
				}
			}
		}
		s.idx++
	}

	return nodes
}

func (s *selector) isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func (s *selector) isWord(c byte) bool {
	return !s.isSpace(c) && c != '[' && c != ']'
}

func (s *selector) next(str string, pos int) byte {
	if pos >= len(str) {
		return 0
	}
	return str[pos+1]
}

func (s *selector) parseExpr(node *selectorNode, str string) {
	// return c >= '0' && c <= '9'
	// 解析形如 a=1,b=2,c="343214 12341" 的表达式
	var key, value bytes.Buffer
	var inQuote bool
	var quoteChar byte

	var i int
	for ; i < len(str); i++ {
		c := str[i]

		// 处理引号
		if c == '"' || c == '\'' {
			if !inQuote {
				inQuote = true
				quoteChar = c
				continue
			} else if quoteChar == c {
				inQuote = false
				node.setProp(key.String(), value.String())
				key.Reset()
				value.Reset()
				continue
			}
		}

		// 处理非引号状态
		if !inQuote {
			if c == '>' || c == '<' {
				value.WriteByte(c)
				if s.next(str, i) == '=' {
					value.WriteByte('=')
					i++
				}
				continue
			}
			if c == '!' || c == '*' {
				value.WriteByte(c)
				if s.next(str, i) == '=' {
					value.WriteByte('=')
					i++
				}
				continue
			}
			if c == '=' {
				// 切换到值的解析
				value.WriteByte(c)
				continue
			}
			if c == ',' {
				// 完成一个键值对
				if key.Len() > 0 && value.Len() > 0 {
					node.setProp(key.String(), value.String())
				}
				key.Reset()
				value.Reset()
				continue
			}

			// 根据是否在等号前后添加到key或value
			if value.Len() == 0 {
				key.WriteByte(c)
			} else {
				value.WriteByte(c)
			}
		} else {
			// 在引号内直接添加字符
			value.WriteByte(c)
		}
	}

	// 处理最后一个键值对
	if key.Len() > 0 && value.Len() > 0 {
		node.setProp(key.String(), value.String())
	}

	// var buf bytes.Buffer
	// for i := 0; i < len(str); i++ {
	// 	if s.isNumber(str[i]) {
	// 		buf.WriteByte(str[i])
	// 	}
	// }
	// return buf.String()
}

type keyWrapper struct {
	key any
	// keyMatched   map[int]bool
	// propsMatched map[int]int
	keyMatched   bool
	propsMatched int
	matchPos     int // 已经match的位置，从0开始（目标对象的时候=len(nodes)）
}

type picker[T any] struct {
	// src any
	// nodes []*selectorNode
	stack  []*keyWrapper
	nodes  []*selectorNode
	result []T
}

func (p *picker[T]) matchProps(kvMap map[string]any, keyWrapper *keyWrapper) {
	node := p.nodes[keyWrapper.matchPos]
	for k, prop := range node.props {
		if vv, ok := kvMap[k]; ok {
			switch prop.op {
			case opErr:
				return
			case opEqual:
				if toString(vv) != prop.value {
					return
				}
				keyWrapper.propsMatched++
			case opLike:
				if !strings.Contains(toString(vv), prop.value.(string)) {
					return
				}
				keyWrapper.propsMatched++
			case opNot:
				if toString(vv) == prop.value {
					return
				}
				keyWrapper.propsMatched++
			case opGt:
				val, ok := toFloat64(vv)
				if !ok {
					return
				}
				if val <= prop.value.(float64) {
					return
				}
				keyWrapper.propsMatched++
			case opGe:
				val, ok := toFloat64(vv)
				if !ok {
					return
				}
				if val < prop.value.(float64) {
					return
				}
				keyWrapper.propsMatched++
			case opLt:
				val, ok := toFloat64(vv)
				if !ok {
					return
				}
				if val >= prop.value.(float64) {
					return
				}
				keyWrapper.propsMatched++
			case opLe:
				val, ok := toFloat64(vv)
				if !ok {
					return
				}
				if val > prop.value.(float64) {
					return
				}
				keyWrapper.propsMatched++
			}
		}
	}
}

// func (p *picker[T]) checkAllPropsMatched() bool {
// 	var pos = -1
// 	for _, keyWrapper := range p.stack {
// 		if keyWrapper.keyMatched[pos+1] && keyWrapper.propsMatched[pos+1] >= len(p.nodes[pos+1].props) {
// 			pos++
// 			if pos == len(p.nodes)-1 {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

func (p *picker[T]) checkMatchPos(keyWrapper *keyWrapper) bool {
	if len(p.stack) == 0 {
		return false
	}
	pos := p.stack[len(p.stack)-1].matchPos
	if keyWrapper.keyMatched && keyWrapper.propsMatched >= len(p.nodes[pos].props) {
		keyWrapper.matchPos++
		return true
	}
	return false
}

func (p *picker[T]) walk(dest any, kk string) {
	keyWrapper := &keyWrapper{
		key: kk,
		// keyMatched:   make(map[int]bool),
		// propsMatched: make(map[int]int),
	}
	if len(p.stack) == 0 {
		keyWrapper.matchPos = 0
	} else {
		keyWrapper.matchPos = p.stack[len(p.stack)-1].matchPos
	}
	p.stack = append(p.stack, keyWrapper)
	defer func() {
		p.stack = p.stack[:len(p.stack)-1]
	}()
	// 字段是否匹配
	node := p.nodes[keyWrapper.matchPos]
	if strings.EqualFold(node.key, kk) || node.key == "" {
		keyWrapper.keyMatched = true
	}
	var v reflect.Value
	var ok bool
	if v, ok = dest.(reflect.Value); !ok {
		v = reflect.ValueOf(dest)
	}
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	// var ref reflect.Value
	// if v.CanAddr() {
	// 	ref = v.Addr()
	// } else {
	// 	ref = reflect.New(v.Type())
	// 	ref.Elem().Set(v)
	// }
	var kvMap map[string]any
	if v.Kind() == reflect.Map || v.Kind() == reflect.Struct || v.Kind() == reflect.Slice {
		kvMap = make(map[string]any)
	}

	switch v.Kind() {
	case reflect.Map:
		for _, k := range v.MapKeys() {
			kk := k.Interface()
			vv := v.MapIndex(k)
			kStr := toString(kk)
			kvMap[kStr] = vv.Interface()
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			vv := v.Index(i)
			kStr := fmt.Sprintf("%d", i)
			kvMap[kStr] = vv.Interface()
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			vv := v.Field(i)
			if vv.CanInterface() {
				kStr := v.Type().Field(i).Name
				kvMap[kStr] = vv.Interface()
			}
		}
	}

	oldPos := keyWrapper.matchPos
	defer func() {
		keyWrapper.matchPos = oldPos
	}()
	if kvMap != nil {
		// 检查属性是否匹配
		p.matchProps(kvMap, keyWrapper)
		if p.checkMatchPos(keyWrapper) && keyWrapper.matchPos == len(p.nodes) {
			keyWrapper.matchPos--
			p.pushResult(dest)
		}
		// 使用确定性的顺序遍历子节点，保证结果有序
		if v.Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				kStr := fmt.Sprintf("%d", i)
				p.walk(kvMap[kStr], kStr)
			}
		} else {
			keys := make([]string, 0, len(kvMap))
			for k := range kvMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				p.walk(kvMap[k], k)
			}
		}
	} else {
		// 如果命中尾节点
		if p.checkMatchPos(keyWrapper) && keyWrapper.matchPos == len(p.nodes) {
			keyWrapper.matchPos--
			p.pushResult(dest)
		}
		// if keyWrapper.keyMatched[len(p.nodes)-1] && keyWrapper.propsMatched[len(p.nodes)-1] == len(p.nodes[len(p.nodes)-1].props) && p.checkAllPropsMatched() {
		// 	p.pushResult(dest)
		// }
	}
}

func (p *picker[T]) pushResult(dest any) {
	var ret any = dest
	if v, ok := dest.(reflect.Value); ok {
		ret = v.Interface()
	}
	if c, ok := ret.(T); ok {
		p.result = append(p.result, c)
		return
	}
	var c T
	if err := Cast(&c, ret); err == nil {
		p.result = append(p.result, c)
	}
}

func pick[T any](src any, rule string) []T {
	selector := &selector{
		src: rule,
	}
	nodes := selector.parse()

	picker := &picker[T]{
		nodes: nodes,
	}
	picker.walk(src, "")
	return picker.result
}

// 从任意对象中收集元素
func Pick[T any](src any, rules ...any) (result []T) {
	var shouldDistinct = false
	defer func() {
		if shouldDistinct {
			var mp = make(map[any]bool)
			var _result []T
			for _, v := range result {
				if mp[v] {
					continue
				}
				mp[v] = true
				_result = append(_result, v)
			}
			result = _result
		}
	}()
	var selectors []string
	for _, rule := range rules {
		if rule == UseDistinct {
			shouldDistinct = true
			continue
		}
		if v, ok := rule.(string); ok {
			selectors = append(selectors, v)
		}
	}
	if len(selectors) == 0 {
		return nil
	}
	if len(selectors) == 1 {
		result = pick[T](src, selectors[0])
		return
	}

	var g WaitGroup
	var ret = make([][]T, 0, len(selectors))
	var mu sync.Mutex
	for i, selector := range selectors {
		i := i
		selector := selector
		g.Go(func() error {
			res := pick[T](src, selector)
			mu.Lock()
			defer mu.Unlock()
			ret[i] = res
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil
	}
	//var finalResult []T
	for _, r := range ret {
		result = append(result, r...)
	}
	return

	//var g *internal.ThreadPool
	// var results []any
	// var stack []any
	// var stackMap = make(map[any]*int)
	// var nodeMatches = make(map[any]map[int]bool)        // 使用s作为key记录节点匹配情况
	// var nodePropMatches = make(map[any]map[string]bool) // 使用s作为key记录属性匹配情况

	// // 添加一个辅助函数来检查是否构成完整路径
	// checkFullMatch := func() bool {
	// 	if len(stack) == 0 {
	// 		return false
	// 	}

	// 	// 直接检查是否有完整的路径匹配
	// 	// 从后向前搜索匹配
	// 	curNode := len(nodes) - 1

	// 	// 从后向前遍历堆栈
	// 	for i := len(stack) - 1; i >= 0; i-- {
	// 		s := stack[i]
	// 		matches, exists := nodeMatches[s]

	// 		// 如果当前节点匹配了当前选择器
	// 		if exists && matches[curNode] {
	// 			curNode--

	// 			// 已经找到了所有选择器的匹配
	// 			if curNode < 0 {
	// 				return true
	// 			}
	// 		}
	// 	}

	// 	// 打印调试信息
	// 	fmt.Printf("无法找到完整路径，当前匹配到第 %d 个选择器，总共 %d 个选择器\n",
	// 		len(nodes)-curNode-1, len(nodes))

	// 	return false
	// }

	// // 检查是否已经匹配了所有需要的属性
	// // checkAllPropsMatched := func(s any, node *selectorNode) bool {
	// // 	if len(node.props) == 0 {
	// // 		return true
	// // 	}

	// // 	propMap, exists := nodePropMatches[s]
	// // 	if !exists {
	// // 		return false
	// // 	}

	// // 	for _, prop := range node.props {
	// // 		if !propMap[prop.key] {
	// // 			return false
	// // 		}
	// // 	}
	// // 	return true
	// // }

	// var stack []*keyWrapper
	// var walk func(any, any, func(any, any))
	// walk = func(dest any, kk any, fn func(k, v any)) {

	// }

	// var cache = make(map[any]any)
	// walk(src, func(k, v any) {
	// 	cache[k] = v
	// })
	//
	//return picker.result
}

// 将任意值转换为字符串
func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", v)
	default:
		return ""
	}
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	}
	return 0, false
}
