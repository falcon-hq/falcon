package remote

import (
	"crypto/cipher"
	"log"
	"net"

	"github.com/maoxs2/falcon/common"
	aeadconn "github.com/maoxs2/go-aead-conn"
	"github.com/maoxs2/go-socks5"
)

type Remote struct {
	enableSnappy bool
	remoteAddr   string
	chunkSize    int
	*s5RemoteServer
}

func NewRemote(chunkSize int, enableSnappy bool) *Remote {
	conf := &socks5.Config{}
	s5, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	s5Server := &s5RemoteServer{
		Server: s5,
	}

	return &Remote{
		enableSnappy:   enableSnappy,
		chunkSize:      chunkSize,
		s5RemoteServer: s5Server,
	}
}

func (r *Remote) HandleConn(conn net.Conn, aead cipher.AEAD) {
	var cryptoConn net.Conn
	if r.enableSnappy {
		cryptoConn = aeadconn.NewAEADCompressConn(common.GetSeed(), r.chunkSize, conn, aead)
	} else {
		cryptoConn = aeadconn.NewAEADConn(common.GetSeed(), r.chunkSize, conn, aead)
	}

	err := r.serverConn(cryptoConn)
	if err != nil {
		log.Println(err)
	}
}
