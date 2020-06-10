package common_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/maoxs2/falcon/common"
)

func TestGetSeed(t *testing.T) {
	s1 := common.GetSeed(nil)
	time.Sleep(time.Second)
	s2 := common.GetSeed(nil)
	if !bytes.Equal(s1, s2) {
		t.Error("seed should be same in 1sec")
		return
	}

	s3 := common.GetSeed(nil)
	time.Sleep(time.Minute)
	s4 := common.GetSeed(nil)
	if bytes.Equal(s3, s4) {
		t.Error("seed should be different beyond 1min")
		return
	}
}
