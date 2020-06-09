package clench

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
)

type HSResponse struct {
	STATUS []byte
}

func (r *HSResponse) Marshal() []byte {
	LENRAND := make([]byte, 1)
	rand.Read(LENRAND)
	RAND := make([]byte, int(LENRAND[0]))
	rand.Read(RAND)

	return bytes.Join([][]byte{LENRAND, r.STATUS, RAND}, nil)
}

func ReadHSResponse(reader io.Reader) (*HSResponse, error) {
	b := make([]byte, 3)
	_, err := reader.Read(b)
	if err != nil {
		return nil, fmt.Errorf("Failed to read read response: %v", err)
	}

	RAND := make([]byte, int(b[0]))
	_, err = reader.Read(RAND)
	if err != nil {
		return nil, fmt.Errorf("Failed to read read response: %v", err)
	}

	r := &HSResponse{}

	r.STATUS = b[1:3]

	return r, nil
}
