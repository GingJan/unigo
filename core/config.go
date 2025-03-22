package core

import (
	"github.com/gingjan/unigo/util"
	"os"
	"strconv"
)

const (
	TimestampBits = "TIME_STAMP_BITS"
	WorkerIdBits  = "WORKER_ID_BITS"
	SequenceBits  = "SEQUENCE_BITS"
	EpochStr      = "EPOCH_STR"
	//EpochSeconds  = "EPOCH_SECONDS"
	BoostPower    = "BOOST_POWER"
	PaddingFactor = "PADDING_FACTOR"
)

type IDGeneratorConfig struct {
	TimestampBits int
	WorkerIdBits  int
	SequenceBits  int
	//workId        int
	EpochStr      string
	BoostPower    int
	PaddingFactor uint64
}

// 初始化配置
func (u *IDGeneratorConfig) Init(t int, w int, s int, e string) {
	u.InitDefault()
	u.TimestampBits = t
	u.WorkerIdBits = w
	u.SequenceBits = s
	u.EpochStr = e
}

// 初始化默认配置
func (u *IDGeneratorConfig) InitDefault() {
	u.TimestampBits = 29
	u.WorkerIdBits = 20
	u.SequenceBits = 15
	u.EpochStr = "2020-08-20"
	u.BoostPower = 3
	u.PaddingFactor = 70
	err := configFromSystemEnv(u) //从环境变量获取配置
	if err != nil {
		panic(err)
	}
}

// 从环境变量获取配置
func configFromSystemEnv(uc *IDGeneratorConfig) (err error) {
	if timestampBits := os.Getenv(TimestampBits); !util.IsBlank(timestampBits) {
		uc.TimestampBits, err = strconv.Atoi(timestampBits)
	}
	if workerIdBits := os.Getenv(WorkerIdBits); !util.IsBlank(workerIdBits) {
		uc.WorkerIdBits, err = strconv.Atoi(workerIdBits)
	}
	if sequenceBits := os.Getenv(SequenceBits); !util.IsBlank(sequenceBits) {
		uc.SequenceBits, err = strconv.Atoi(sequenceBits)
	}
	if epochStr := os.Getenv(EpochStr); !util.IsBlank(epochStr) {
		uc.EpochStr = epochStr
	}
	if boostPower := os.Getenv(BoostPower); !util.IsBlank(boostPower) {
		uc.BoostPower, err = strconv.Atoi(boostPower)
	}
	if paddingFactor := os.Getenv(PaddingFactor); !util.IsBlank(paddingFactor) {
		uc.PaddingFactor, err = strconv.ParseUint(PaddingFactor, 10, 64)
	}
	return
}
