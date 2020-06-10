package common

import "golang.org/x/crypto/sha3"

// Sha3Sum256 is a helper func to calc sha3.Sum256 and get []byte in one line
func Sha3Sum256(b []byte) []byte {
	hash := sha3.Sum256(b)
	return hash[:]
}
