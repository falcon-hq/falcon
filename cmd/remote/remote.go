package remotecmd

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"net"

	logging "github.com/ipfs/go-log"

	"github.com/falcon-hq/falcon/common"
	aeadconn "github.com/maoxs2/go-aead-conn"
	"github.com/maoxs2/go-socks5"
)

var log = logging.Logger("falcon-remote")

type Remote struct {
	serverPort   int
	key          string
	enableSnappy bool
	remoteAddr   string
	chunkSize    int
	*s5RemoteServer
}

func NewRemote(serverPort int, key string, chunkSize int, enableSnappy bool) *Remote {
	conf := &socks5.Config{}
	s5, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	if len(key) == 0 {
		log.Warn("key is empty")
	}

	return &Remote{
		serverPort:   serverPort,
		key:          key,
		enableSnappy: enableSnappy,
		chunkSize:    chunkSize,
		s5RemoteServer: &s5RemoteServer{
			Server: s5,
		},
	}
}

func (r *Remote) MainLoop() {
	serveAddr := fmt.Sprintf(":%d", r.serverPort)
	l, err := net.Listen("tcp", serveAddr)
	if err != nil {
		panic(err)
	}

	log.Infof("remote server is serving on %s, key: %s \n", serveAddr, r.key)
	log.Infof("You can run falcon with the following arguments in your local device: \n ./falcon local -k %s -r [Remote_Server_IP]:%d -chunk %d \n", r.key, r.serverPort, r.chunkSize)
	block, err := aes.NewCipher(common.Sha3Sum256([]byte(r.key)))
	if err != nil {
		panic(err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Errorf("failed accept the conn from client: %s", err)
			continue
		}

		log.Infof("accepted conn from %s", conn.RemoteAddr())
		go r.HandleConn(conn, aead)
	}
}

func (r *Remote) HandleConn(conn net.Conn, aead cipher.AEAD) {
	var cryptoConn net.Conn
	if r.enableSnappy {
		cryptoConn = aeadconn.NewAEADCompressConn(common.GetSeed([]byte(r.key)), r.chunkSize, conn, aead)
	} else {
		cryptoConn = aeadconn.NewAEADConn(common.GetSeed([]byte(r.key)), r.chunkSize, conn, aead)
	}

	err := r.serverConn(cryptoConn)
	if err != nil {
		log.Error(err)
	}
}
