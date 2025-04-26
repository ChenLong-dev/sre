package sw

import (
	"container/ring"
	"sync"
	"time"
)

// 基于桶的滑动窗口
type SlidingWindow struct {
	// 环形队列
	bucketRing *ring.Ring
	// 单个桶的时间间隔
	bucketInterval time.Duration
	// 桶读写锁
	bucketRWMutex sync.RWMutex
	// 上次访问时间
	lastAccessTime time.Time
}

// 新建滑动窗口
func NewSlidingWindow(bucketInterval time.Duration, bucketCount int) *SlidingWindow {
	buckets := ring.New(bucketCount)
	for i := 0; i < buckets.Len(); i++ {
		buckets.Value = new(Bucket)
		buckets = buckets.Next()
	}

	return &SlidingWindow{
		bucketRing:     buckets,
		bucketInterval: bucketInterval,
		lastAccessTime: time.Now(),
	}
}

// 失败
func (w *SlidingWindow) Fail() {
	w.bucketRWMutex.Lock()
	b := w.getCurrentBucket()
	b.Fail()
	w.bucketRWMutex.Unlock()
}

// 成功
func (w *SlidingWindow) Success() {
	w.bucketRWMutex.Lock()
	b := w.getCurrentBucket()
	b.Success()
	w.bucketRWMutex.Unlock()
}

// 增加计数
func (w *SlidingWindow) Increase() {
	w.bucketRWMutex.Lock()
	b := w.getCurrentBucket()
	b.Increase()
	w.bucketRWMutex.Unlock()
}

// 减少计数
func (w *SlidingWindow) Decrease() {
	w.bucketRWMutex.Lock()
	b := w.getCurrentBucket()
	b.Decrease()
	w.bucketRWMutex.Unlock()
}

// 重置
func (w *SlidingWindow) Reset() {
	w.bucketRWMutex.Lock()

	w.bucketRing.Do(func(x interface{}) {
		x.(*Bucket).Reset()
	})

	w.bucketRWMutex.Unlock()
}

// 滑动
func (w *SlidingWindow) Slide() {
	w.bucketRWMutex.Lock()
	w.getCurrentBucket()
	w.bucketRWMutex.Unlock()
}

// 获取当前的bucket
func (w *SlidingWindow) getCurrentBucket() *Bucket {
	currentBucket := w.bucketRing.Value.(*Bucket)

	timeDiff := time.Now().Sub(w.lastAccessTime)

	if timeDiff > w.bucketInterval {
		// 滑动并重置时间差内的桶
		for i := 0; i < w.bucketRing.Len(); i++ {
			w.bucketRing = w.bucketRing.Next()
			currentBucket = w.bucketRing.Value.(*Bucket)
			currentBucket.Reset()

			timeDiff = time.Duration(int64(timeDiff) - int64(w.bucketInterval))
			if timeDiff < w.bucketInterval {
				break
			}
		}

		w.lastAccessTime = time.Now()
	}

	return currentBucket
}

// 失败次数
func (w *SlidingWindow) FailureCount() int64 {
	w.bucketRWMutex.RLock()

	var failures int64
	w.bucketRing.Do(func(x interface{}) {
		b := x.(*Bucket)
		failures += b.failure
	})

	w.bucketRWMutex.RUnlock()

	return failures
}

// 成功次数
func (w *SlidingWindow) SuccessCount() int64 {
	w.bucketRWMutex.RLock()

	var successes int64
	w.bucketRing.Do(func(x interface{}) {
		b := x.(*Bucket)
		successes += b.success
	})

	w.bucketRWMutex.RUnlock()

	return successes
}

// 错误比例
func (w *SlidingWindow) ErrorRate() float64 {
	var total int64
	var failures int64

	w.bucketRWMutex.RLock()

	w.bucketRing.Do(func(x interface{}) {
		b := x.(*Bucket)
		total += b.failure + b.success
		failures += b.failure
	})

	w.bucketRWMutex.RUnlock()

	if total == 0 {
		return 0.0
	}

	return float64(failures) / float64(total)
}

// 次数
func (w *SlidingWindow) Count() int64 {
	w.bucketRWMutex.RLock()

	var count int64
	w.bucketRing.Do(func(x interface{}) {
		b := x.(*Bucket)
		count += b.count
	})

	w.bucketRWMutex.RUnlock()

	return count
}
