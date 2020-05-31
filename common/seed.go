package common

import (
	"encoding/binary"
	"time"
)

func GetSeed() []byte {
	return Sha3Sum256(packInt64(getTimestampAboutMinute()))
}

func getTimestampAboutMinute() int64 {
	return time.Now().Unix() << 6
}

func packInt64(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}
