package core

import (
	"fmt"
	"github.com/gingjan/unigo/util"
	"sync"
)

type IDGenerator struct {
	ringbuffer Rb
	bita       *BitsAllocator
	/** Customer epoch, unit as second. For example 2016-05-20 (ms: 1463673600000)*/
	//epochStr     string
	epochSeconds uint64
	workId       uint64
	boostPower   int
	config       *IDGeneratorConfig
	initOnce     sync.Once
}

func (c *IDGenerator) InitDefaultWithWorkDB(w *MachineNodeDBConfig, ringBufferVersion int) {
	c.initOnce.Do(func() {
		cfg := &IDGeneratorConfig{}
		//初始化配置
		cfg.InitDefault()
		c.InitIDGlobalCfgAndWorkerDB(cfg, w, ringBufferVersion)
	})
}

func (c *IDGenerator) InitIDGlobalCfgAndWorkerDB(cfg *IDGeneratorConfig, w *MachineNodeDBConfig, ringBufferVersion int) {
	c.initOnce.Do(func() {
		c.config = cfg
		workId, _ := NewMachineNode(w)
		c.bita = &BitsAllocator{}
		//初始化BitsAllocator
		c.bita.Init(c.config.TimestampBits, c.config.WorkerIdBits, c.config.SequenceBits)
		c.workId = uint64(workId)
		if c.workId > c.bita.MaxWorkerId {
			panic(fmt.Sprintf("Worker id %v exceeds the max %v", c.workId, c.bita.MaxWorkerId))
		}
		c.epochSeconds = util.DateToSecond(c.config.EpochStr)
		c.boostPower = c.config.BoostPower
		bufferSize := (c.bita.MaxSequence + 1) << c.boostPower
		if ringBufferVersion == 2 {
			c.ringbuffer = InitRingBufferV2(bufferSize, c.config.PaddingFactor)
		} else {
			c.ringbuffer = InitRingBuffer(bufferSize, c.config.PaddingFactor)
		}
		c.ringbuffer.SetIDGeneFunc(c.geneIDsForOneSecond)
		c.ringbuffer.AsyncPadding() //初始化id池，一次性填满
	})
}

func (c *IDGenerator) GetID() (id uint64, err error) {
	id, err = c.ringbuffer.Take()
	return
}

func (c *IDGenerator) MustGetID() (id uint64, err error) {
	for true {
		if id, err = c.ringbuffer.Take(); err == nil {
			return
		}
	}
	return
}

// 解析ID
func (c *IDGenerator) ParseID(id uint64) string {
	sequence := (id << (c.bita.TotalBits - c.bita.SequenceBits)) >> (c.bita.TotalBits - c.bita.SequenceBits)
	workerId := (id << (c.bita.TimestampBits)) >> (c.bita.TotalBits - c.bita.WorkerIdBits)
	deltaSeconds := id >> (c.bita.WorkerIdBits + c.bita.SequenceBits)
	return fmt.Sprintf("{\"ID\":\"%v\",\"deltaSeconds\":\"%v\",\"workerId\":\"%v\",\"sequence\":\"%v\"}",
		id, c.epochSeconds+deltaSeconds, workerId, sequence)
}

// 本秒内的所有序列号，
func (c *IDGenerator) geneIDsForOneSecond(currentSecond uint64) (idList []uint64) {
	// Initialize result list size of (max sequence + 1)
	listSize := c.bita.MaxSequence + 1
	//本秒的第一个序列号
	firstSeqID := c.bita.Allocate(currentSecond-c.epochSeconds, c.workId, 0)
	var offset uint64
	for offset = 0; offset < listSize; offset++ {
		idList = append(idList, firstSeqID+offset)
	}
	return idList
}
