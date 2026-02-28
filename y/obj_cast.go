// objx.go
package y

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Converter 类型转换器
type Converter struct {
	tagPriority []string // 标签优先级
	mu          sync.RWMutex
	fieldCache  map[reflect.Type]map[string]fieldInfo // 字段名和索引缓存
}

var (
	defaultConverter = NewConverter()
)

// NewConverter 创建新的类型转换器
func NewConverter() *Converter {
	return &Converter{
		tagPriority: []string{"db", "json"},
		fieldCache:  make(map[reflect.Type]map[string]fieldInfo),
	}
}

// Convert 类型转换入口函数
func Cast(dest, src interface{}) error {
	return defaultConverter.Convert(dest, src)
}

// Convert 执行类型转换
func (c *Converter) Convert(dest, src interface{}) error {
	if dest == nil {
		return fmt.Errorf("destination cannot be nil")
	}

	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		return fmt.Errorf("destination must be a non-nil pointer")
	}

	return c.convertValue(destVal.Elem(), reflect.ValueOf(src))
}

// fieldInfo 存储字段的索引和名称信息
type fieldInfo struct {
	Index  int    // 字段索引
	Name   string // 字段名
	Format string // 时间格式化标签（用于 time.Time -> string 转换）
}

const defaultTimeFormat = "2006-01-02 15:04:05"

// javaFormatToGo 将 Java/常见风格的时间格式转换为 Go 的 time 格式
func javaFormatToGo(format string) string {
	replacer := strings.NewReplacer(
		"yyyy", "2006",
		"yy", "06",
		"MM", "01",
		"dd", "02",
		"HH", "15",
		"hh", "03",
		"mm", "04",
		"ss", "05",
		"SSS", "000",
		"SS", "00",
		"S", "0",
	)
	return replacer.Replace(format)
}

// getTimeFormat 获取 Go 风格的时间格式字符串，如果已是 Go 格式则直接返回
func getTimeFormat(format string) string {
	if format == "" {
		return defaultTimeFormat
	}
	if strings.Contains(format, "2006") || strings.Contains(format, "06") {
		return format
	}
	return javaFormatToGo(format)
}

// 获取字段名和索引映射
func (c *Converter) getFieldMap(t reflect.Type) map[string]fieldInfo {
	c.mu.RLock()
	if fields, ok := c.fieldCache[t]; ok {
		c.mu.RUnlock()
		return fields
	}
	c.mu.RUnlock()

	fieldMap := make(map[string]fieldInfo)
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldName := field.Name

			// 检查标签
			var tagName string
			for _, tag := range c.tagPriority {
				if tagValue := field.Tag.Get(tag); tagValue != "" {
					if idx := strings.Index(tagValue, ","); idx != -1 {
						tagValue = tagValue[:idx]
					}
					if tagValue != "" && tagValue != "-" {
						tagName = tagValue
						break
					}
				}
			}

		formatTag := field.Tag.Get("format")

		info := fieldInfo{
			Index:  i,
			Name:   fieldName,
			Format: formatTag,
		}

			if tagName != "" {
				fieldMap[tagName] = info
			}
			// 始终添加原始字段名映射
			fieldMap[fieldName] = info

			// 处理嵌套结构体
			if field.Type.Kind() == reflect.Struct && field.Anonymous {
				nestedMap := c.getFieldMap(field.Type)
				for k, v := range nestedMap {
					// 避免覆盖外层字段
					if _, exists := fieldMap[k]; !exists {
						fieldMap[k] = fieldInfo{
							Index: i,
							Name:  fieldName + "." + v.Name,
						}
					}
				}
			}
		}
	}

	c.mu.Lock()
	c.fieldCache[t] = fieldMap
	c.mu.Unlock()

	return fieldMap
}

// 转换值
func (c *Converter) convertValue(dest, src reflect.Value) error {
	// 处理nil源
	if !src.IsValid() {
		dest.Set(reflect.Zero(dest.Type()))
		return nil
	}

	// 处理Data类型
	if dest.Type() == reflect.TypeOf(Data[interface{}]{}) {
		return c.convertToData(dest, src)
	}

	// 处理指针
	if dest.Kind() == reflect.Ptr {
		return c.convertToPtr(dest, src)
	}

	src = reflect.Indirect(src)
	destType := dest.Type()

	// 类型完全匹配
	if src.Type().AssignableTo(destType) {
		dest.Set(src)
		return nil
	}

	// 处理time.Time类型的转换（包括指针类型）
	if destType == reflect.TypeOf(time.Time{}) || destType == reflect.TypeOf(&time.Time{}) {
		// 处理接口类型，获取实际值
		actualSrc := src
		for actualSrc.Kind() == reflect.Interface && !actualSrc.IsNil() {
			actualSrc = actualSrc.Elem()
		}
		switch actualSrc.Kind() {
		case reflect.String:
			// 使用Guess函数尝试解析各种时间格式
			t, err := Guess(actualSrc.String())
			if err == nil {
				if destType.Kind() == reflect.Ptr {
					// 处理*time.Time类型
					dest.Set(reflect.ValueOf(&t))
				} else {
					// 处理time.Time类型
					dest.Set(reflect.ValueOf(t))
				}
			} else {
				// 解析失败则返回零值，不报错
				if destType.Kind() == reflect.Ptr {
					dest.Set(reflect.Zero(destType)) // nil for *time.Time
				} else {
					dest.Set(reflect.ValueOf(time.Time{}))
				}
			}
			return nil
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			// 处理时间戳（秒）
			t := time.Unix(actualSrc.Int(), 0)
			if destType.Kind() == reflect.Ptr {
				// 处理*time.Time类型
				dest.Set(reflect.ValueOf(&t))
			} else {
				// 处理time.Time类型
				dest.Set(reflect.ValueOf(t))
			}
			return nil
		}
		// 其他类型不处理，返回零值
		if destType.Kind() == reflect.Ptr {
			dest.Set(reflect.Zero(destType)) // nil for *time.Time
		} else {
			dest.Set(reflect.ValueOf(time.Time{}))
		}
		return nil
	}

	// 处理不同类型的转换
	switch dest.Kind() {
	case reflect.Struct:
		return c.convertToStruct(dest, src)
	case reflect.Slice, reflect.Array:
		return c.convertToSlice(dest, src)
	case reflect.Map:
		return c.convertToMap(dest, src)
	case reflect.String:
		return c.convertToString(dest, src)
	case reflect.Bool:
		return c.convertToBool(dest, src)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 处理time.Time到整型的转换
		if src.Type().ConvertibleTo(reflect.TypeOf(time.Time{})) {
			t := src.Interface().(time.Time)
			dest.SetInt(t.Unix())
			return nil
		}
		return c.convertToInt(dest, src)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 处理time.Time到无符号整型的转换
		if src.Type().ConvertibleTo(reflect.TypeOf(time.Time{})) {
			t := src.Interface().(time.Time)
			dest.SetUint(uint64(t.Unix()))
			return nil
		}
		return c.convertToUint(dest, src)
	case reflect.Float32, reflect.Float64:
		// 处理time.Time到浮点型的转换
		if src.Type().ConvertibleTo(reflect.TypeOf(time.Time{})) {
			t := src.Interface().(time.Time)
			dest.SetFloat(float64(t.UnixNano()) / 1e9)
			return nil
		}
		return c.convertToFloat(dest, src)
	case reflect.Interface:
		dest.Set(src)
		return nil
	}

	return fmt.Errorf("unsupported conversion from %v to %v", src.Type(), destType)
}

// 转换为Data类型
func (c *Converter) convertToData(dest, src reflect.Value) error {
	// 获取Data的泛型类型
	dataType := dest.Type()
	var data Data[any]

	// 创建新的Data实例
	if dataType == reflect.TypeOf(Data[any]{}) {
		data = make(Data[any])
	} else {
		// 处理具体的泛型Data类型
		newData := reflect.MakeMap(dataType)
		data = newData.Interface().(Data[any])
	}

	// 设置Data实例到目标
	dest.Set(reflect.ValueOf(data))

	src = reflect.Indirect(src)
	if !src.IsValid() {
		return nil
	}

	// 获取泛型类型信息
	var genericType reflect.Type
	if dest.Type().Kind() == reflect.Map {
		if elemType := dest.Type().Elem(); elemType.Kind() == reflect.Interface {
			// 非泛型Data[any]情况
			genericType = nil
		} else {
			// 获取泛型参数类型
			genericType = dest.Type().Elem()
		}
	}

	switch src.Kind() {
	case reflect.Map:
		iter := src.MapRange()
		for iter.Next() {
			key := fmt.Sprint(iter.Key().Interface())
			if key == dataKey {
				continue // 跳过保留键
			}

			if genericType != nil && genericType.Kind() != reflect.Interface {
				// 创建目标类型的值并转换
				val := reflect.New(genericType).Elem()
				if err := c.convertValue(val, iter.Value()); err != nil {
					return fmt.Errorf("failed to convert field %s: %w", key, err)
				}
				data[key] = val.Interface()
			} else {
				data[key] = iter.Value().Interface()
			}
		}

	case reflect.Struct:
		srcType := src.Type()
		fieldMap := c.getFieldMap(srcType)

		// 处理结构体字段
		for _, info := range fieldMap {
			fieldName := info.Name
			fieldValue := src.Field(info.Index)

			// 处理嵌套字段
			if strings.Contains(fieldName, ".") {
				parts := strings.Split(fieldName, ".")
				current := data
				for i, part := range parts[:len(parts)-1] {
					if _, exists := current[part]; !exists {
						current[part] = make(Data[any])
					}
					var ok bool
					if current, ok = current[part].(Data[any]); !ok {
						// 如果类型不是Data[any]，则跳过
						break
					}
					if i == len(parts)-2 {
						// 最后一个部分，设置值
						current[parts[len(parts)-1]] = fieldValue.Interface()
					}
				}
				continue
			}

			if genericType != nil && genericType.Kind() != reflect.Interface {
				// 创建目标类型的值并转换
				val := reflect.New(genericType).Elem()
				if err := c.convertValue(val, fieldValue); err != nil {
					return fmt.Errorf("failed to convert field %s: %w", fieldName, err)
				}
				data[fieldName] = val.Interface()
			} else {
				data[fieldName] = fieldValue.Interface()
			}
		}

	default:
		// 其他类型直接存储到dataKey下
		if genericType != nil && genericType.Kind() != reflect.Interface {
			val := reflect.New(genericType).Elem()
			if err := c.convertValue(val, src); err != nil {
				return fmt.Errorf("failed to convert value: %w", err)
			}
			data[dataKey] = val.Interface()
		} else {
			data[dataKey] = src.Interface()
		}
	}

	return nil
}

// 转换为指针
func (c *Converter) convertToPtr(dest, src reflect.Value) error {
	// 处理源为nil的情况
	if !src.IsValid() || (src.Kind() == reflect.Ptr && src.IsNil()) {
		dest.Set(reflect.Zero(dest.Type()))
		return nil
	}

	// 解引用指针或接口
	if src.Kind() == reflect.Ptr || src.Kind() == reflect.Interface {
		src = src.Elem()
		// 解引用后再次检查是否有效（处理 interface{} 包含 nil 的情况）
		if !src.IsValid() {
			dest.Set(reflect.Zero(dest.Type()))
			return nil
		}
	}

	// 如果目标为nil，创建新实例
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}

	// 递归转换值
	return c.convertValue(dest.Elem(), src)
}

// 设置结构体字段值
func (c *Converter) setStructField(dest reflect.Value, fieldInfo fieldInfo, srcVal reflect.Value) error {
	// 处理嵌套字段 (如 "User.Name")
	fieldNames := strings.Split(fieldInfo.Name, ".")
	field := dest

	// 遍历嵌套字段路径
	for i, name := range fieldNames {
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				if !field.CanSet() {
					// 无法设置的字段直接跳过，不报错
					return nil
				}
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}

		// 获取当前层级的字段
		currentField := field.FieldByName(name)
		if !currentField.IsValid() {
			// 不存在的字段直接跳过，不报错
			return nil
		}

		// 如果是最后一个字段，设置值
		if i == len(fieldNames)-1 {
			if !currentField.CanSet() {
				// 不可设置的字段直接跳过，不报错
				return nil
			}

			// 特殊处理time.Time类型（包括指针类型）
			fieldType := currentField.Type()
			if fieldType == reflect.TypeOf(time.Time{}) || fieldType == reflect.TypeOf(&time.Time{}) {
				// 处理接口类型，获取实际值
				actualSrcVal := srcVal
				for actualSrcVal.Kind() == reflect.Interface && !actualSrcVal.IsNil() {
					actualSrcVal = actualSrcVal.Elem()
				}
				switch actualSrcVal.Kind() {
				case reflect.String:
					// 使用Guess函数尝试解析各种时间格式
					t, err := Guess(actualSrcVal.String())
					if err == nil {
						if fieldType.Kind() == reflect.Ptr {
							// 处理*time.Time类型
							currentField.Set(reflect.ValueOf(&t))
						} else {
							// 处理time.Time类型
							currentField.Set(reflect.ValueOf(t))
						}
					}
					// 解析失败则跳过，不报错
					return nil
				case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
					// 处理时间戳（秒）
					t := time.Unix(actualSrcVal.Int(), 0)
					if fieldType.Kind() == reflect.Ptr {
						// 处理*time.Time类型
						currentField.Set(reflect.ValueOf(&t))
					} else {
						// 处理time.Time类型
						currentField.Set(reflect.ValueOf(t))
					}
					return nil
				}
			}

			// 特殊处理 time.Time -> string（带 format 标签）
			if fieldType.Kind() == reflect.String {
				actualSrcVal := srcVal
				for actualSrcVal.Kind() == reflect.Interface && !actualSrcVal.IsNil() {
					actualSrcVal = actualSrcVal.Elem()
				}
				actualSrcVal = reflect.Indirect(actualSrcVal)
				if actualSrcVal.IsValid() && actualSrcVal.Type() == reflect.TypeOf(time.Time{}) {
					t := actualSrcVal.Interface().(time.Time)
					currentField.SetString(t.Format(getTimeFormat(fieldInfo.Format)))
					return nil
				}
			}

			// 其他类型正常转换
			return c.convertValue(currentField, srcVal)
		}

		// 移动到下一个嵌套字段
		field = currentField
	}

	return nil
}

// 转换为结构体
func (c *Converter) convertToStruct(dest, src reflect.Value) error {
	src = reflect.Indirect(src)
	if !src.IsValid() {
		return nil
	}

	destType := dest.Type()
	fieldMap := c.getFieldMap(destType)

	// 处理Data类型
	if dest.CanAddr() {
		if _, ok := dest.Addr().Interface().(*Data[any]); ok {
			return c.convertToData(dest, src)
		}
	}

	switch src.Kind() {
	case reflect.Map:
		if src.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("map key must be string, got %v", src.Type().Key())
		}

		iter := src.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			fi, ok := fieldMap[key]
			if !ok {
				continue
			}

			field := dest.Field(fi.Index)
			fieldType := field.Type()

			// 特殊处理 time.Time 类型（包括指针类型）
			if fieldType == reflect.TypeOf(time.Time{}) || fieldType == reflect.TypeOf(&time.Time{}) {
				srcVal := iter.Value()
				for srcVal.Kind() == reflect.Interface && !srcVal.IsNil() {
					srcVal = srcVal.Elem()
				}
				if srcVal.Kind() == reflect.String {
					t, err := Guess(srcVal.String())
					if err == nil && field.CanSet() {
						if fieldType.Kind() == reflect.Ptr {
							field.Set(reflect.ValueOf(&t))
						} else {
							field.Set(reflect.ValueOf(t))
						}
					}
					continue
				}
			}

			// 特殊处理 time.Time -> string（带 format 标签）
			if fieldType.Kind() == reflect.String {
				srcVal := iter.Value()
				for srcVal.Kind() == reflect.Interface && !srcVal.IsNil() {
					srcVal = srcVal.Elem()
				}
				srcVal = reflect.Indirect(srcVal)
				if srcVal.IsValid() && srcVal.Type() == reflect.TypeOf(time.Time{}) {
					t := srcVal.Interface().(time.Time)
					field.SetString(t.Format(getTimeFormat(fi.Format)))
					continue
				}
			}

			if err := c.setStructField(dest, fi, iter.Value()); err != nil {
				return fmt.Errorf("field %s: %w", key, err)
			}
		}
		return nil

	case reflect.Struct:
		srcType := src.Type()
		srcFieldMap := c.getFieldMap(srcType)

		for destKey, destInfo := range fieldMap {
			srcInfo, exists := srcFieldMap[destKey]
			if !exists {
				continue
			}

			srcField := src.Field(srcInfo.Index)
			if !srcField.IsValid() {
				continue
			}

			field := dest.Field(destInfo.Index)
			fieldType := field.Type()

			// 特殊处理 time.Time 类型（包括指针类型）
			if fieldType == reflect.TypeOf(time.Time{}) || fieldType == reflect.TypeOf(&time.Time{}) {
				srcVal := srcField
				for srcVal.Kind() == reflect.Interface && !srcVal.IsNil() {
					srcVal = srcVal.Elem()
				}
				if srcVal.Kind() == reflect.String {
					t, err := Guess(srcVal.String())
					if err == nil && field.CanSet() {
						if fieldType.Kind() == reflect.Ptr {
							field.Set(reflect.ValueOf(&t))
						} else {
							field.Set(reflect.ValueOf(t))
						}
					}
					continue
				}
			}

			// 特殊处理 time.Time -> string（带 format 标签）
			if fieldType.Kind() == reflect.String {
				srcVal := srcField
				for srcVal.Kind() == reflect.Interface && !srcVal.IsNil() {
					srcVal = srcVal.Elem()
				}
				srcVal = reflect.Indirect(srcVal)
				if srcVal.IsValid() && srcVal.Type() == reflect.TypeOf(time.Time{}) {
					t := srcVal.Interface().(time.Time)
					field.SetString(t.Format(getTimeFormat(destInfo.Format)))
					continue
				}
			}

			if err := c.setStructField(dest, destInfo, srcField); err != nil {
				return fmt.Errorf("field %s: %w", destKey, err)
			}
		}
		return nil

	case reflect.Interface, reflect.Ptr:
		// 处理接口和指针类型
		if src.IsNil() {
			return nil
		}
		return c.convertToStruct(dest, src.Elem())

	default:
		// 尝试将基本类型转换为单字段结构体
		if dest.NumField() == 1 {
			field := dest.Type().Field(0)
			if field.IsExported() {
				return c.convertValue(dest.Field(0), src)
			}
		}
		return fmt.Errorf("cannot convert %v to struct", src.Type())
	}
}

// 转换为切片
func (c *Converter) convertToSlice(dest, src reflect.Value) error {
	src = reflect.Indirect(src)
	if !src.IsValid() {
		dest.Set(reflect.MakeSlice(dest.Type(), 0, 0))
		return nil
	}

	var sliceLen int
	switch src.Kind() {
	case reflect.Slice, reflect.Array:
		sliceLen = src.Len()
		if sliceLen == 0 {
			dest.Set(reflect.MakeSlice(dest.Type(), 0, 0))
			return nil
		}
	default:
		sliceLen = 1
	}

	destType := dest.Type()
	elemType := destType.Elem()
	slice := reflect.MakeSlice(destType, sliceLen, sliceLen)

	switch src.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < src.Len(); i++ {
			elem := reflect.New(elemType).Elem()
			if err := c.convertValue(elem, src.Index(i)); err != nil {
				return fmt.Errorf("element %d: %w", i, err)
			}
			slice.Index(i).Set(elem)
		}

	default:
		elem := reflect.New(elemType).Elem()
		if err := c.convertValue(elem, src); err != nil {
			return fmt.Errorf("element 0: %w", err)
		}
		slice.Index(0).Set(elem)
	}

	dest.Set(slice)
	return nil
}

// 转换为映射
func (c *Converter) convertToMap(dest, src reflect.Value) error {
	src = reflect.Indirect(src)
	if !src.IsValid() {
		dest.Set(reflect.MakeMap(dest.Type()))
		return nil
	}

	destType := dest.Type()
	keyType := destType.Key()
	elemType := destType.Elem()
	m := reflect.MakeMap(destType)

	switch src.Kind() {
	case reflect.Map:
		iter := src.MapRange()
		for iter.Next() {
			key := reflect.New(keyType).Elem()
			if err := c.convertValue(key, iter.Key()); err != nil {
				return fmt.Errorf("map key: %w", err)
			}

			value := reflect.New(elemType).Elem()
			if err := c.convertValue(value, iter.Value()); err != nil {
				return fmt.Errorf("map value for key %v: %w", key, err)
			}

			m.SetMapIndex(key, value)
		}

	case reflect.Struct:
		srcType := src.Type()
		for i := 0; i < srcType.NumField(); i++ {
			field := srcType.Field(i)
			if !field.IsExported() {
				continue
			}

			key := reflect.ValueOf(field.Name)
			if key.Type() != keyType {
				tmpKey := reflect.New(keyType).Elem()
				if err := c.convertValue(tmpKey, key); err != nil {
					return fmt.Errorf("map key for field %s: %w", field.Name, err)
				}
				key = tmpKey
			}

			value := reflect.New(elemType).Elem()
			if err := c.convertValue(value, src.Field(i)); err != nil {
				return fmt.Errorf("map value for field %s: %w", field.Name, err)
			}

			m.SetMapIndex(key, value)
		}

	default:
		return fmt.Errorf("cannot convert %v to map", src.Type())
	}

	dest.Set(m)
	return nil
}

// 转换为字符串
func (c *Converter) convertToString(dest, src reflect.Value) error {
	return c.convertToStringWithFormat(dest, src, "")
}

// convertToStringWithFormat 转换为字符串，支持 format 参数（用于 time.Time 格式化）
func (c *Converter) convertToStringWithFormat(dest, src reflect.Value, format string) error {
	src = reflect.Indirect(src)
	if !src.IsValid() {
		dest.SetString("")
		return nil
	}

	// 处理 time.Time -> string
	if src.Type() == reflect.TypeOf(time.Time{}) {
		t := src.Interface().(time.Time)
		dest.SetString(t.Format(getTimeFormat(format)))
		return nil
	}

	switch src.Kind() {
	case reflect.String:
		dest.SetString(src.String())
	case reflect.Bool:
		dest.SetString(fmt.Sprintf("%v", src.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dest.SetString(fmt.Sprintf("%d", src.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dest.SetString(fmt.Sprintf("%d", src.Uint()))
	case reflect.Float32, reflect.Float64:
		dest.SetString(fmt.Sprintf("%v", src.Float()))
	case reflect.Struct:
		return fmt.Errorf("cannot convert %v to string", src.Type())
	case reflect.Interface, reflect.Ptr:
		return c.convertToStringWithFormat(dest, src.Elem(), format)
	default:
		return fmt.Errorf("cannot convert %v to string", src.Type())
	}
	return nil
}

// 转换为布尔值
func (c *Converter) convertToBool(dest, src reflect.Value) error {
	src = reflect.Indirect(src)
	if !src.IsValid() {
		dest.SetBool(false)
		return nil
	}

	var b bool
	switch src.Kind() {
	case reflect.Bool:
		b = src.Bool()
	case reflect.String:
		s := strings.TrimSpace(strings.ToLower(src.String()))
		if s == "" || s == "false" || s == "0" || s == "no" || s == "off" {
			b = false
		} else {
			b = true
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b = src.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b = src.Uint() != 0
	case reflect.Float32, reflect.Float64:
		b = src.Float() != 0
	case reflect.Interface, reflect.Ptr:
		return c.convertToBool(dest, src.Elem())
	default:
		return fmt.Errorf("cannot convert %v to bool", src.Type())
	}

	dest.SetBool(b)
	return nil
}

// 转换为整数
func (c *Converter) convertToInt(dest, src reflect.Value) error {
	src = reflect.Indirect(src)
	if !src.IsValid() {
		dest.SetInt(0)
		return nil
	}

	var i int64
	switch src.Kind() {
	case reflect.Bool:
		if src.Bool() {
			i = 1
		}
	case reflect.String:
		s := strings.TrimSpace(src.String())
		if s == "" {
			i = 0
		} else {
			val, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse %q as int: %w", s, err)
			}
			i = val
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i = src.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i = int64(src.Uint())
	case reflect.Float32, reflect.Float64:
		i = int64(src.Float())
	case reflect.Interface, reflect.Ptr:
		return c.convertToInt(dest, src.Elem())
	default:
		return fmt.Errorf("cannot convert %v to int", src.Type())
	}

	dest.SetInt(i)
	return nil
}

// 转换为无符号整数
func (c *Converter) convertToUint(dest, src reflect.Value) error {
	src = reflect.Indirect(src)
	if !src.IsValid() {
		dest.SetUint(0)
		return nil
	}

	var u uint64
	switch src.Kind() {
	case reflect.Bool:
		if src.Bool() {
			u = 1
		}
	case reflect.String:
		s := strings.TrimSpace(src.String())
		if s == "" {
			u = 0
		} else {
			val, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse %q as uint: %w", s, err)
			}
			u = val
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := src.Int()
		if i < 0 {
			return fmt.Errorf("cannot convert negative value %d to uint", i)
		}
		u = uint64(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u = src.Uint()
	case reflect.Float32, reflect.Float64:
		f := src.Float()
		if f < 0 {
			return fmt.Errorf("cannot convert negative value %v to uint", f)
		}
		u = uint64(f)
	case reflect.Interface, reflect.Ptr:
		return c.convertToUint(dest, src.Elem())
	default:
		return fmt.Errorf("cannot convert %v to uint", src.Type())
	}

	dest.SetUint(u)
	return nil
}

// 转换为浮点数
func (c *Converter) convertToFloat(dest, src reflect.Value) error {
	src = reflect.Indirect(src)
	if !src.IsValid() {
		dest.SetFloat(0)
		return nil
	}

	var f float64
	switch src.Kind() {
	case reflect.Bool:
		if src.Bool() {
			f = 1
		}
	case reflect.String:
		s := strings.TrimSpace(src.String())
		if s == "" {
			f = 0
		} else {
			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("cannot parse %q as float: %w", s, err)
			}
			f = val
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f = float64(src.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f = float64(src.Uint())
	case reflect.Float32, reflect.Float64:
		f = src.Float()
	case reflect.Interface, reflect.Ptr:
		return c.convertToFloat(dest, src.Elem())
	default:
		return fmt.Errorf("cannot convert %v to float", src.Type())
	}

	dest.SetFloat(f)
	return nil
}
