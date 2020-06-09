package clench_test

import (
	"bytes"
	"testing"

	"github.com/maoxs2/falcon/clench"
)

func TestRes(t *testing.T) {
	res := &clench.HSResponse{
		STATUS: []byte{0, 0},
	}
	res2, err := clench.ReadHSResponse(bytes.NewBuffer(res.Marshal()))
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(res.STATUS, res2.STATUS) {
		t.Errorf("%v should be same to %v", res, res2)
	}
}
