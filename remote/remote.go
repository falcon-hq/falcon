package remote

import (
	"crypto/cipher"
	"log"
	"net"

	"github.com/maoxs2/falcon/clench"
	"github.com/maoxs2/falcon/common"
	aeadconn "github.com/maoxs2/go-aead-conn"
)

type Remote struct {
	enableSnappy bool
	// remoteAddr   string
	chunkSize int
	*clench.Server
}

func NewRemote(chunkSize int, enableSnappy bool) *Remote {
	return &Remote{
		enableSnappy: enableSnappy,
		chunkSize:    chunkSize,
		Server:       &clench.Server{},
	}
}

func (r *Remote) HandleConn(conn net.Conn, aead cipher.AEAD) {
	var cryptoConn net.Conn
	if r.enableSnappy {
		cryptoConn = aeadconn.NewAEADCompressConn(common.GetSeed(), r.chunkSize, conn, aead)
	} else {
		cryptoConn = aeadconn.NewAEADConn(common.GetSeed(), r.chunkSize, conn, aead)
	}

	err := r.Server.HandleConn(cryptoConn)
	if err != nil {
		log.Println(err)
	}
}
