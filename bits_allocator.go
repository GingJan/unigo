package unigo

import "sync"

type BitsAllocator struct {
	TotalBits      int // 总共位数 64bit
	timestampShift int //时间戳的左移位数
	workerIdShift  int //机器码的左移位数

	TimestampBits int //时间戳bit数
	WorkerIdBits  int //机器码bit数
	SequenceBits  int //序列号bit数

	maxDeltaSeconds uint64 //时间戳最大值
	MaxWorkerId     uint64 //机器码最大值
	MaxSequence     uint64 //序列号最大值
	initOnce        sync.Once
}

// 初始化
// timestampBits 时间戳bit数
// workerIdBits 机器码bit数
// sequenceBits 序列号bit数
func (b *BitsAllocator) Init(timestampBits int, workerIdBits int, sequenceBits int) {
	b.initOnce.Do(func() {
		if timestampBits+workerIdBits+sequenceBits < 64 {
			panic("Less than 64 bits")
		}
		if timestampBits+workerIdBits+sequenceBits > 64 {
			panic("more than 64 bits")
		}
		//initialize bits
		b.TimestampBits = timestampBits
		b.WorkerIdBits = workerIdBits
		b.SequenceBits = sequenceBits

		//初始化最大值
		var m int64
		m = -1
		b.maxDeltaSeconds = uint64(^(m << timestampBits)) // 等价 (m << timestampBits) - 1
		b.MaxWorkerId = uint64(^(m << workerIdBits))
		b.MaxSequence = uint64(^(m << sequenceBits))

		// 初始化偏移量
		b.timestampShift = workerIdBits + sequenceBits
		b.workerIdShift = sequenceBits
		b.TotalBits = 1 << 6 //64
	})
}

func (b *BitsAllocator) Allocate(deltaSeconds uint64, workerId uint64, sequence uint64) uint64 {
	return (deltaSeconds << b.timestampShift) | (workerId << b.workerIdShift) | sequence
}
