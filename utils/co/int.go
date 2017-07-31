package co

import "sync/atomic"

// Int64 协程安全的int64
type Int64 int64

// Add 增加
func (i *Int64) Add(n int64) int64 {
	return atomic.AddInt64((*int64)(i), n)
}

// Get 减少
func (i *Int64) Get() int64 {
	return atomic.LoadInt64((*int64)(i))
}

// Int32 协程安全的int32
type Int32 int32

// Add 增加
func (i *Int32) Add(n int32) int32 {
	return atomic.AddInt32((*int32)(i), n)
}

// Get 减少
func (i *Int32) Get() int32 {
	return atomic.LoadInt32((*int32)(i))
}
