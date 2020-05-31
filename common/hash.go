package common

import "golang.org/x/crypto/sha3"

func Sha3Sum256(b []byte) []byte {
	hash := sha3.Sum256(b)
	return hash[:]
}
