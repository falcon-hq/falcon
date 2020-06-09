package clench

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
)

type connWriteReader interface {
	io.Writer
	io.Reader
	RemoteAddr() net.Addr
}

type HSRequest struct {
	PROTOCOL byte
	PORT     int

	LENDST int
	DST    []byte

	destAddr *AddrSpec
}

func (r *HSRequest) Marshal() []byte {
	LENRAND := make([]byte, 1)
	b := make([]byte, 5+len(r.DST)+int(LENRAND[0]))

	rand.Read(LENRAND)
	RAND := make([]byte, int(LENRAND[0]))
	rand.Read(RAND)
	b[0] = byte(LENRAND[0])
	b[1] = r.PROTOCOL
	binary.BigEndian.PutUint16(b[2:4], uint16(r.PORT))
	b[4] = byte(r.LENDST)

	copy(b[5:], r.DST)
	copy(b[5+len(r.DST):], RAND)

	return b
}

func ReadHSRequest(reader io.Reader) (*HSRequest, error) {
	// Read the version byte
	b := make([]byte, 5)
	_, err := reader.Read(b)
	if err != nil {
		return nil, fmt.Errorf("Failed to get fixed 5 bytes header: %v", err)
	}

	r := &HSRequest{}

	r.PROTOCOL = b[1]
	r.PORT = int(binary.BigEndian.Uint16(b[2:4]))
	r.LENDST = int(b[4])

	r.DST = make([]byte, r.LENDST)
	_, err = reader.Read(r.DST)
	if err != nil {
		return nil, fmt.Errorf("Failed to get DST: %v", err)
	}

	rand := make([]byte, int(b[0]))
	_, _ = reader.Read(rand)

	r.destAddr, err = NewAddrSpec(r.PROTOCOL&0b1 == 1, r.DST, r.PORT)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func NewHSRequest(protocol byte, port int, dst []byte) (*HSRequest, error) {
	r := &HSRequest{}

	r.PROTOCOL = protocol
	r.PORT = port
	r.LENDST = len(dst)
	r.DST = dst

	var err error
	r.destAddr, err = NewAddrSpec(r.PROTOCOL&0b1 == 1, r.DST, r.PORT)
	if err != nil {
		return nil, err
	}

	return r, nil
}
