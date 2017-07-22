package co

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Pair 键值对
type Pair interface {
	Key() interface{}
	Value() interface{}
	Format(layout string) string
}

// Comap 一个线程安全的库
type Comap interface {
	// 获取键值
	Get(interface{}) (interface{}, bool)
	// 设置键值
	Set(interface{}, interface{})
	// 删除键值,并返回值
	Remove(interface{}) interface{}
	// 删除符合条件的键值,并返回删除的数量
	RemoveMatch(func(Pair) bool) int
	// 清空
	Clear()
	// 遍历
	Foreach(func(interface{}, interface{}))
	// 复制
	Clone() map[interface{}]interface{}
	// 键数组
	Keys() []interface{}
	// 排序后的键组
	SortedKeys(func(Pair, Pair) bool) []interface{}
	// 值数组
	Values() []interface{}
	// 排序后的值组
	SortedValues(func(Pair, Pair) bool) []interface{}
	// 长度
	Len() int
	// 包括键
	Has(interface{}) bool
	// 键值对
	Pairs() []Pair
	// 排序后的键值组
	SortedPairs(func(Pair, Pair) bool) []Pair
	// 格式化输出
	Format(string) string
}

// ComapImpl comap的实现
type ComapImpl struct {
	mutex sync.RWMutex
	kvs   map[interface{}]interface{}
}

// Get 获取键值
func (impl *ComapImpl) Get(k interface{}) (interface{}, bool) {
	impl.mutex.RLock()
	v, exist := impl.kvs[k]
	impl.mutex.RUnlock()
	return v, exist
}

// Format 格式化输出
func (impl *ComapImpl) Format(layout string) string {
	pairs := impl.Pairs()

	//FIXME:性能不好，需要优化
	s := "{"
	for index, kv := range pairs {
		s += kv.Format(layout)
		if index != len(pairs)-1 {
			s += " "
		}
	}

	return s + "}"
}

type pairImpl struct {
	K interface{} `json:"key"`
	V interface{} `json:"value"`
}

func (pair *pairImpl) Key() interface{} {
	return pair.K
}

func (pair *pairImpl) Value() interface{} {
	return pair.V
}

func (pair *pairImpl) String() string {
	b, _ := json.Marshal(pair)
	return string(b)
}

func (pair *pairImpl) Format(layout string) string {
	k := strings.Replace(layout, "%value", fmt.Sprintf("%v", pair.Value()), -1)
	k = strings.Replace(k, "%key", fmt.Sprintf("%v", pair.Key()), -1)
	return k
}

// Remove 删除键值,并返回值
func (impl *ComapImpl) Remove(k interface{}) interface{} {
	impl.mutex.Lock()
	defer impl.mutex.Unlock()
	if impl.kvs == nil {
		return nil
	}

	v := impl.kvs[k]
	delete(impl.kvs, k)

	return v
}

// Clear 清空
func (impl *ComapImpl) Clear() {
	impl.mutex.Lock()
	defer impl.mutex.Unlock()
	impl.kvs = map[interface{}]interface{}{}
}

// RemoveMatch 删除符合条件的键值,并返回删除的数量
func (impl *ComapImpl) RemoveMatch(f func(Pair) bool) int {
	impl.mutex.RLock()
	defer impl.mutex.Unlock()
	i := 0
	for key, value := range impl.kvs {
		matched := f(&pairImpl{key, value})
		if matched {
			i++
			delete(impl.kvs, key)
		}
	}
	return i
}

// Set 设置键值
func (impl *ComapImpl) Set(k, v interface{}) {
	impl.mutex.Lock()
	impl.kvs[k] = v
	impl.mutex.Unlock()
}

// Foreach 遍历
func (impl *ComapImpl) Foreach(f func(interface{}, interface{})) {
	impl.mutex.RLock()
	defer impl.mutex.Unlock()
	for key, value := range impl.kvs {
		f(key, value)
	}
}

// Clone 复制
func (impl *ComapImpl) Clone() map[interface{}]interface{} {
	impl.mutex.RLock()
	newKvs := make(map[interface{}]interface{}, len(impl.kvs))
	for key, value := range impl.kvs {
		newKvs[key] = value
	}
	impl.mutex.RUnlock()

	return newKvs
}

// Keys 键数组
func (impl *ComapImpl) Keys() []interface{} {
	impl.mutex.RLock()
	keys := make([]interface{}, len(impl.kvs))
	i := 0
	for key, _ := range impl.kvs {
		keys[i] = key
		i++
	}
	impl.mutex.RUnlock()

	return keys
}

// Values 值数组
func (impl *ComapImpl) Values() []interface{} {
	impl.mutex.RLock()
	values := make([]interface{}, len(impl.kvs))
	i := 0
	for _, value := range impl.kvs {
		values[i] = value
		i++
	}
	impl.mutex.RUnlock()

	return values
}

// Len 长度
func (impl *ComapImpl) Len() int {
	return len(impl.kvs)
}

// Has 包括键
func (impl *ComapImpl) Has(k interface{}) bool {
	impl.mutex.RLock()
	_, exist := impl.kvs[k]
	impl.mutex.RUnlock()
	return exist
}

// Pairs 键值对
func (impl *ComapImpl) Pairs() []Pair {
	impl.mutex.RLock()
	pairs := make([]Pair, 0, len(impl.kvs))
	for key, value := range impl.kvs {
		pairs = append(pairs, &pairImpl{key, value})
	}
	impl.mutex.RUnlock()

	return pairs
}

// SortedValues 排序后的键值
func (impl *ComapImpl) SortedValues(f func(Pair, Pair) bool) []interface{} {
	pairs := impl.SortedPairs(f)
	values := make([]interface{}, len(pairs))
	for i, p := range pairs {
		values[i] = p.Value()
	}

	return values
}

// SortedKeys 排序后的键组
func (impl *ComapImpl) SortedKeys(f func(Pair, Pair) bool) []interface{} {
	pairs := impl.SortedPairs(f)
	keys := make([]interface{}, len(pairs))
	for i, p := range pairs {
		keys[i] = p.Key()
	}

	return keys
}

// SortedPairs 排序后的键值组
func (impl *ComapImpl) SortedPairs(f func(Pair, Pair) bool) []Pair {
	pairs := impl.Pairs()
	sort.Slice(pairs, func(i, j int) bool {
		return f(pairs[i], pairs[j])
	})

	return pairs
}

// NewFrom 从特定类型创建
func NewFrom(interface{}) Comap {
	// TODO:如果是个数组
	// TODO:如果是个map
	// TODO:如果是个对象
	// TODO:如果是其他
	return nil
}

// NewBy 仅需要封装一层安全操作
func NewBy(source map[interface{}]interface{}) Comap {
	return &ComapImpl{
		kvs: source,
	}
}

// New 创建一个新的
func New() Comap {
	return &ComapImpl{
		kvs: make(map[interface{}]interface{}, 3),
	}
}
