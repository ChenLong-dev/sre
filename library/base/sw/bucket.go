package sw

// 计数桶
type Bucket struct {
	// 失败次数
	failure int64
	// 成功次数
	success int64

	// 次数
	count int64
}

// 重置
func (b *Bucket) Reset() {
	b.failure = 0
	b.success = 0
	b.count = 0
}

// 失败
func (b *Bucket) Fail() {
	b.failure++
}

// 成功
func (b *Bucket) Success() {
	b.success++
}

// 增加计数
func (b *Bucket) Increase() {
	b.count++
}

// 减少计数
func (b *Bucket) Decrease() {
	b.count--
}
