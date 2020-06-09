package clench_test

import (
	"bytes"
	"testing"

	"github.com/maoxs2/falcon/clench"
)

func TestReq(t *testing.T) {
	req, err := clench.NewHSRequest(0b1, 443, []byte("baidu.com"))
	if err != nil {
		t.Error(err)
		return
	}
	req2, err := clench.ReadHSRequest(bytes.NewBuffer(req.Marshal()))
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(req.DST, req2.DST) {
		t.Errorf("%v is not %v", req, req2)
	}
	if req.PORT != req2.PORT {
		t.Errorf("%v is not %v", req, req2)
	}
	if req.PROTOCOL != req2.PROTOCOL {
		t.Errorf("%v is not %v", req, req2)
	}
}
