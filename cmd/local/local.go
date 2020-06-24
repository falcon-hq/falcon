package localcmd

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"net"

	"github.com/falcon-hq/falcon/auth"
	"github.com/spf13/viper"

	"github.com/falcon-hq/falcon/common"
	logging "github.com/ipfs/go-log"
	aeadconn "github.com/maoxs2/go-aead-conn"
)

var log = logging.Logger("falcon-local")

type Local struct {
	socks5Addr   string
	enableSnappy bool
	remoteAddr   string
	chunkSize    int
	authMan      *auth.Authenticator
	key          string
}

func NewLocal(socks5Addr, remoteAddr, key string, chunkSize int, authUser, authPass string, enableSnappy bool) *Local {
	var authMan *auth.Authenticator
	if len(authUser) != 0 {
		authMan = auth.NewAuthenticator(map[string]string{
			authUser: authPass,
		})
	} else {
		authMan = auth.NewAuthenticator(nil)
	}

	if len(key) == 0 {
		log.Warn("key is empty")
	}

	return &Local{
		socks5Addr:   socks5Addr,
		enableSnappy: enableSnappy,
		remoteAddr:   remoteAddr,
		key:          key,
		chunkSize:    chunkSize,
		authMan:      authMan,
	}
}

func (l *Local) MainLoop() {
	listener, err := net.Listen("tcp", l.socks5Addr)
	if err != nil {
		panic(err)
	}

	log.Infof("serving on local: %s \n", l.socks5Addr)
	log.Infof("local connects with remote %s with key %s \n", l.remoteAddr, l.key)
	block, err := aes.NewCipher(common.Sha3Sum256([]byte(viper.GetString("key"))))
	if err != nil {
		panic(err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	for {
		localConn, err := listener.Accept()
		if err != nil {
			log.Errorf("failed to accept local conn: %s", err)
			continue
		}

		version := []byte{0}
		if _, err := localConn.Read(version); err != nil {
			log.Errorf("socks: Failed to get version byte: %v", err)
		}

		if version[0] != 0x05 {
			err := fmt.Errorf("unsupported socks version: %v", version)
			log.Errorf("socks version error: %v", err)
		}

		if err := l.authMan.Auth(localConn, localConn); err != nil {
			log.Errorf("auth failed: %v", err)
			continue
		}

		log.Info("new app conn")
		go l.HandleConn(localConn, aead)
	}
}

func (l *Local) HandleConn(lc net.Conn, aead cipher.AEAD) {
	externalConn, err := net.Dial("tcp", l.remoteAddr)
	if err != nil {
		log.Errorf("failed connecting to remote server: %s: %s", l.remoteAddr, err)
		return
	}

	var cryptoConn net.Conn
	if l.enableSnappy {
		cryptoConn = aeadconn.NewAEADCompressConn(common.GetSeed([]byte(l.key)), l.chunkSize, externalConn, aead)
	} else {
		cryptoConn = aeadconn.NewAEADConn(common.GetSeed([]byte(l.key)), l.chunkSize, externalConn, aead)
	}

	// lc -> sc
	go func() {
		buf := make([]byte, l.chunkSize)
		for {
			n, err := lc.Read(buf)
			if err == io.EOF {
				externalConn.(*net.TCPConn).CloseWrite() // todo
				break
			}
			if err != nil {
				log.Errorf("outgoing error: %v", err)
				break
			}

			if n == 0 {
				continue
			}

			log.Debugf("sending: %x", buf[:n])
			cryptoConn.Write(buf[:n])
		}
	}()

	// sc -> lc
	go func() {
		buf := make([]byte, l.chunkSize)
		for {
			n, err := cryptoConn.Read(buf)
			if err == io.EOF {
				lc.(*net.TCPConn).CloseWrite()
				break
			}
			if err != nil {
				log.Errorf("incoming error: %v", err)
				break
			}

			if n == 0 {
				continue
			}

			log.Debugf("recving: %x", buf[:n])
			lc.Write(buf[:n])
		}
	}()
}
