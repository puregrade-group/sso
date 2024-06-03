package snowflake

import (
	"fmt"
	"sync"
	"time"
)

type Snowflake struct {
	timestamp int64
	nodeID    uint16
	sequence  uint16
	mu        sync.Mutex
}

const (
	epoch        int64  = 1577836800000 // 01.01.2020
	nodeBits     uint8  = 11
	sequenceBits uint8  = 12
	maxNodeID    uint16 = (1 << nodeBits) - 1
	maxSequence  uint16 = (1 << sequenceBits) - 1
	timeShift           = nodeBits + sequenceBits
	nodeShift           = sequenceBits
)

func NewSnowflake(nodeID uint16) (*Snowflake, error) {
	if nodeID > maxNodeID {
		return nil, fmt.Errorf("nodeID can't be greater then %d", maxNodeID)
	}

	return &Snowflake{
		timestamp: 0,
		nodeID:    nodeID,
		sequence:  0,
	}, nil
}

func (sf *Snowflake) Generate() uint64 {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	now := time.Now().UnixNano() / 1e6
	if now < epoch {
		panic("Clock moved backwards")
	}

	if now == sf.timestamp {
		sf.sequence++
		if sf.sequence > maxSequence {
			for now <= sf.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		sf.sequence = 0
	}

	sf.timestamp = now

	id := uint64((now-epoch)<<timeShift) | (uint64(sf.nodeID) << nodeShift) | uint64(sf.sequence)
	return id
}
