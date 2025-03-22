package unigo

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type RingBuffer struct {
	bufferSize       uint64
	indexMask        int64
	paddingThreshold int64 //至少填充的id个数
	lastSecond       uint64
	running          uint64
	tail             int64
	cursor           int64
	slots            []uint64
	flags            []uint32
	mu               sync.Mutex
	idGeneFunc       func(uint64) []uint64
}

// 初始化ringbuffer
func InitRingBuffer(bufferSize uint64, paddingFactor uint64) *RingBuffer {
	r := &RingBuffer{
		bufferSize: bufferSize,
		indexMask:  (int64)(bufferSize - 1),
		slots:      make([]uint64, bufferSize),
		flags:      make([]uint32, bufferSize),
		//slots:            make([]atomic.Value, bufferSize),
		//flags:            make([]atomic.Value, bufferSize),
		paddingThreshold: (int64)(bufferSize * paddingFactor / 100),
	}
	var i uint64
	for i = 0; i < bufferSize; i++ {
		r.flags[i] = CAN_PUT_FLAG
	}
	r.tail = START_POINT
	r.cursor = START_POINT
	r.lastSecond = uint64(time.Now().Unix())
	return r
}

func (r *RingBuffer) SetIDGeneFunc(geneFunc func(uint64) []uint64) {
	r.idGeneFunc = geneFunc
}

// 把uid放到缓冲区里，等待后续使用
func (r *RingBuffer) Put(uid uint64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentTail := atomic.LoadInt64(&r.tail)
	currentCursor := atomic.LoadInt64(&r.cursor)
	distance := currentTail - currentCursor
	if distance == r.indexMask {
		//到达最大buffersize，拒绝再放入uid
		fmt.Printf("Rejected putting buffer for uid:%v,tail:%v,cursor:%v\n", uid, currentTail, currentCursor)
		return false
	}

	nextTailIndex := r.calSlotIndex(currentTail + 1)
	if atomic.LoadUint32(&r.flags[nextTailIndex]) != CAN_PUT_FLAG {
		//当前位置的标志不是可put，拒绝
		fmt.Printf("Curosr not in can put status,rejected uid:%v,tail:%v,cursor:%v\n", uid, currentTail, currentCursor)
		return false
	}

	atomic.StoreUint64(&r.slots[nextTailIndex], uid)
	atomic.StoreUint32(&r.flags[nextTailIndex], CAN_TAKE_FLAG)
	atomic.AddInt64(&r.tail, 1)
	return true
}

// 从id池里取一个id
func (r *RingBuffer) Take() (uint64, error) {
	currentCursor := atomic.LoadInt64(&r.cursor)
	nextCursor := Uint64UpdateAndGet(&r.cursor, func(o int64) int64 {
		if o == atomic.LoadInt64(&r.tail) {
			return o
		} else {
			return o + 1
		}
	})
	if nextCursor < currentCursor {
		panic("Curosr can't move back")
	}

	currentTail := atomic.LoadInt64(&r.tail)
	if currentTail-nextCursor < r.paddingThreshold {
		go r.AsyncPadding() //异步填充
	}
	if currentTail == currentCursor {
		//拒绝
		return 0, errors.New("Rejected take uid")
	}
	nextCursorIndex := r.calSlotIndex(nextCursor)
	if atomic.LoadUint32(&r.flags[nextCursorIndex]) != CAN_TAKE_FLAG {
		return 0, errors.New("Curosr not in can take status")
	}
	uid := atomic.LoadUint64(&r.slots[nextCursorIndex])
	atomic.StoreUint32(&r.flags[nextCursorIndex], CAN_PUT_FLAG)
	return uid, nil
}

// 异步预生成id，并填充id池
func (r *RingBuffer) AsyncPadding() {
	// 判断填充id的逻辑是否在运行中
	if !atomic.CompareAndSwapUint64(&r.running, NOT_RUNNING, RUNNING) {
		return
	}
	defer atomic.CompareAndSwapUint64(&r.running, RUNNING, NOT_RUNNING)

	isFullRingBuffer := false
	// fill the rest slots until to catch the cursor
	for !isFullRingBuffer {
		//生成本秒内的所有序列号
		uidList := r.idGeneFunc(atomic.AddUint64(&r.lastSecond, 1)) //当前秒 在 上一秒的基础上+1秒
		for _, uid := range uidList {
			isFullRingBuffer = !r.Put(uid)
			if isFullRingBuffer { //id池满了
				break
			}
		}
	}
}

func (r *RingBuffer) calSlotIndex(sequence int64) int64 {
	//return sequence % r.indexMask
	return sequence & r.indexMask
}
