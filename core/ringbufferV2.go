package core

import (
	"errors"
	"fmt"
	"github.com/gingjan/unigo/util"
	"sync"
	"sync/atomic"
	"time"
)

type PaddedTailStruct struct {
	tail int64
	_    CacheLinePad
}

type PaddedCursorStruct struct {
	cursor int64
	_      CacheLinePad
}

type PaddedRunStruct struct {
	running uint64
	_       CacheLinePad
}

type PaddedSlotStruct struct {
	slot uint64
	_    CacheLinePad
}

type PaddedflagStruct struct {
	flag uint64
	_    CacheLinePad
}

type CacheLinePad struct {
	_ [CacheLinePadSizeV2]byte
}

const CacheLinePadSizeV2 = 32

type RingBufferV2 struct {
	bufferSize       uint64
	indexMask        int64
	paddingThreshold int64 //至少填充的id个数
	lastSecond       uint64
	running          PaddedRunStruct
	tail             PaddedTailStruct //缓冲区尾部
	cursor           PaddedCursorStruct
	slots            []PaddedSlotStruct
	flags            []PaddedflagStruct
	mu               sync.Mutex
	idGeneFunc       func(uint64) []uint64 //id生成函数
}

// 初始化ringbuffer
func InitRingBufferV2(bufferSize uint64, paddingFactor uint64) *RingBufferV2 {
	r := &RingBufferV2{
		bufferSize: bufferSize, //缓冲区大小
		indexMask:  (int64)(bufferSize - 1),
		slots:      make([]PaddedSlotStruct, bufferSize),
		flags:      make([]PaddedflagStruct, bufferSize),
		//slots:            make([]atomic.Value, bufferSize),
		//flags:            make([]atomic.Value, bufferSize),
		paddingThreshold: (int64)(bufferSize * paddingFactor / 100),
	}
	var i uint64
	for i = 0; i < bufferSize; i++ {
		r.flags[i].flag = CAN_PUT_FLAGV2
	}
	r.tail = PaddedTailStruct{
		tail: -1,
	}
	r.cursor = PaddedCursorStruct{
		cursor: -1,
	}
	r.lastSecond = uint64(time.Now().Unix())
	return r
}

func (r *RingBufferV2) SetIDGeneFunc(geneFunc func(uint64) []uint64) {
	r.idGeneFunc = geneFunc
}

// 把uid放到缓冲区里，等待后续使用
func (r *RingBufferV2) Put(uid uint64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentTail := atomic.LoadInt64(&r.tail.tail) //
	currentCursor := atomic.LoadInt64(&r.cursor.cursor)
	distance := currentTail - currentCursor
	if distance == r.indexMask {
		//到达最大buffersize，拒绝再放入uid
		fmt.Printf("Rejected putting buffer for uid:%v,tail:%v,cursor:%v\n", uid, currentTail, currentCursor)
		return false
	}

	nextTailIndex := r.calSlotIndex(currentTail + 1)
	if atomic.LoadUint64(&r.flags[nextTailIndex].flag) != CAN_PUT_FLAGV2 {
		//当前位置的标志不是可put，拒绝
		fmt.Printf("Curosr not in can put status,rejected uid:%v,tail:%v,cursor:%v\n", uid, currentTail, currentCursor)
		return false
	}

	atomic.StoreUint64(&r.slots[nextTailIndex].slot, uid)             //把新预生成的id放入，等待后续使用
	atomic.StoreUint64(&r.flags[nextTailIndex].flag, CAN_TAKE_FLAGV2) //标记该id的可用状态
	atomic.AddInt64(&r.tail.tail, 1)                                  //尾部往后移动1
	return true
}

// 从id池里取一个id
func (r *RingBufferV2) Take() (uint64, error) {
	currentCursor := atomic.LoadInt64(&r.cursor.cursor)
	nextCursor := util.Uint64UpdateAndGet(&r.cursor.cursor, func(o int64) int64 {
		if o == atomic.LoadInt64(&r.tail.tail) {
			return o
		} else {
			return o + 1
		}
	})
	if nextCursor < currentCursor {
		panic("Curosr can't move back")
	}

	currentTail := atomic.LoadInt64(&r.tail.tail)    //尾部
	if currentTail-nextCursor < r.paddingThreshold { //预填充个数不够
		//异步填充
		go r.AsyncPadding()
	}
	if currentTail == currentCursor {
		//拒绝
		return 0, errors.New("Rejected take uid")
	}
	nextCursorIndex := r.calSlotIndex(nextCursor)
	if atomic.LoadUint64(&r.flags[nextCursorIndex].flag) != CAN_TAKE_FLAGV2 {
		return 0, errors.New("Curosr not in can take status")
	}
	uid := atomic.LoadUint64(&r.slots[nextCursorIndex].slot)
	atomic.StoreUint64(&r.flags[nextCursorIndex].flag, CAN_PUT_FLAGV2)
	return uid, nil
}

// 异步预生成id，并填充id池
func (r *RingBufferV2) AsyncPadding() {
	// 判断填充id的逻辑是否在运行中
	if !atomic.CompareAndSwapUint64(&r.running.running, NOT_RUNNING, RUNNING) {
		return
	}
	defer atomic.CompareAndSwapUint64(&r.running.running, RUNNING, NOT_RUNNING)

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

func (r *RingBufferV2) calSlotIndex(sequence int64) int64 {
	//return sequence % r.indexMask
	return sequence & r.indexMask
}
