package util

import (
	"fmt"
	"sync"
	"time"
)

var (
	Sf    *Snowflake
	epoch int64
)

// Snowflake 结构体定义
type Snowflake struct {
	mu        sync.Mutex
	timestamp int64 // 时间戳占用的位数
	workerID  int64 // 机器ID占用的位数
	sequence  int64 // 序列号占用的位数
}

const (
	workerBits         = 8 //10
	sequenceBits       = 9 //12
	maxWorkerID        = -1 ^ (-1 << workerBits)
	maxSequence        = -1 ^ (-1 << sequenceBits)
	timeLeftShift      = sequenceBits + workerBits
	workerLeftShift    = sequenceBits
	timestampLeftShift = sequenceBits + workerBits + 1
)

// 初始化函数
func init() {
	epoch = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / 1e6

	// 在其他地方可以直接使用sf生成唯一ID
	Sf, _ = NewSnowflake(1)

}

// NewSnowflake 创建一个Snowflake实例
func NewSnowflake(workerID int64) (*Snowflake, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, fmt.Errorf("worker ID must be between 0 and %d", maxWorkerID)
	}
	return &Snowflake{
		timestamp: 0,
		workerID:  workerID,
		sequence:  0,
	}, nil
}

// GenerateID 生成唯一ID
func (s *Snowflake) GenerateID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentTime := time.Now().UnixNano() / 1e6

	if currentTime == s.timestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			for currentTime <= s.timestamp {
				currentTime = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		s.sequence = 0
	}

	if currentTime < s.timestamp {
		return 0, fmt.Errorf("clock is moving backwards")
	}

	s.timestamp = currentTime
	ID := ((currentTime - epoch) << timestampLeftShift) |
		(s.workerID << workerLeftShift) |
		s.sequence

	return ID, nil
}
