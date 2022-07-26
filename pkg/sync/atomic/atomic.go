package atomic

import "sync/atomic"

// 封装用于并发确认 bool 的值

type Boolean uint32 // 定义一个 Boolean
// Get 获取 Boolean 的值
func (b *Boolean) Get() bool {
	return atomic.LoadUint32((*uint32)(b)) != 0
}

// Set 修改 Boolean 的值
func (b *Boolean) Set(v bool) {
	if v {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}
