package local

import (
	"crypto/cipher"
	"io"
	"log"
	"net"

	"github.com/maoxs2/falcon/common"
	aeadconn "github.com/maoxs2/go-aead-conn"
	"github.com/maoxs2/go-socks5"
)

type Local struct {
	enableSnappy bool
	remoteAddr   string
	chunkSize    int
	*s5LocalServer
}

func NewLocal(remoteAddr string, chunkSize int, authUser, authPass string, enableSnappy bool) *Local {
	conf := &socks5.Config{}
	if len(authUser) > 0 {
		var c socks5.StaticCredentials
		c[authUser] = authPass
		conf.AuthMethods = append(conf.AuthMethods, &socks5.UserPassAuthenticator{
			Credentials: c,
		})
	}

	s5, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	s5Server := &s5LocalServer{
		Server: s5,
	}

	return &Local{
		enableSnappy:  enableSnappy,
		remoteAddr:    remoteAddr,
		chunkSize:     chunkSize,
		s5LocalServer: s5Server,
	}
}

func (l *Local) HandleConn(lc net.Conn, aead cipher.AEAD) {
	externalConn, err := net.Dial("tcp", l.remoteAddr)
	if err != nil {
		log.Printf("failed connecting to remote server: %s: %s", l.remoteAddr, err)
	}

	var cryptoConn net.Conn
	if l.enableSnappy {
		cryptoConn = aeadconn.NewAEADCompressConn(common.GetSeed(), l.chunkSize, externalConn, aead)
	} else {
		cryptoConn = aeadconn.NewAEADConn(common.GetSeed(), l.chunkSize, externalConn, aead)
	}

	// lc -> sc
	go func() {
		buf := make([]byte, 256)
		for {
			l, err := lc.Read(buf)
			if err == io.EOF {
				cryptoConn.Close()
				break
			}
			if err != nil {
				log.Println(err)
				break
			}

			if l == 0 {
				continue
			}

			log.Printf("sending: %x", buf[:l])
			cryptoConn.Write(buf[:l])
		}
	}()

	// sc -> lc
	go func() {
		buf := make([]byte, 256)
		for {
			l, err := cryptoConn.Read(buf)
			if err == io.EOF {
				lc.Close()
				break
			}
			if err != nil {
				log.Println(err)
				break
			}

			if l == 0 {
				continue
			}

			log.Printf("recving: %x", buf[:l])
			lc.Write(buf[:l])
		}
	}()
}
