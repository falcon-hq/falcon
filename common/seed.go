package common

import (
	"encoding/binary"
	"time"
)

// GetSeed returns a time-interval-guranteed seed based on the timestamp
func GetSeed(salt []byte) []byte {
	ts := packInt64(getTimestampAboutMinute())
	tsAndHash := make([]byte, 0, 8+32)
	tsAndHash = append(ts, Sha3Sum256(salt)...)
	return Sha3Sum256(tsAndHash)
}

func getTimestampAboutMinute() int64 {
	return time.Now().Unix() >> 6
}

func packInt64(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}
