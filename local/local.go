package local

import (
	"bytes"
	"crypto/cipher"
	"io"
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/maoxs2/falcon/clench"
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

	id := uuid.New().String()

	raw, err := S5ReqConvert(lc)
	if err != nil {
		log.Printf("convert %s err: %v", id, err)
	}
	cryptoConn.Write(raw)

	raw, err = S5ResConvert(cryptoConn)
	if err != nil {
		log.Printf("convert %s err: %v", id, err)
	}
	lc.Write(raw)

	// lc -> sc
	go func() {

		buf := make([]byte, 1024)
		for {
			l, err := lc.Read(buf)
			if err == io.EOF {
				log.Printf("local conn EOF, closing remote")
				externalConn.(*net.TCPConn).CloseWrite()
				break
			}
			if err != nil {
				log.Printf("conn %s err: %v", id, err)
				break
			}

			if l == 0 {
				continue
			}

			log.Printf("conn %s sending: %x", id, buf[:l])
			cryptoConn.Write(buf[:l])
		}
	}()

	// sc -> lc
	go func() {
		buf := make([]byte, 1024)
		for {
			l, err := cryptoConn.Read(buf)
			if err == io.EOF {
				log.Printf("remote conn EOF, closing local")
				lc.Close()
				break
			}
			if err != nil {
				log.Printf("conn %s err: %v", id, err)
				break
			}

			if l == 0 {
				continue
			}

			log.Printf("conn %s recving: %x", id, buf[:l])
			lc.Write(buf[:l])
		}
	}()
}

func S5ReqConvert(reader io.Reader) ([]byte, error) {
	r, err := socks5.NewRequest(reader)
	if err != nil {
		return nil, err
	}

	protocol := byte(0)
	if r.Command == 0x03 {
		protocol += 0b10
	}

	if len(r.DestAddr.FQDN) > 0 {
		log.Printf("req %s", r.DestAddr.Address())
		protocol += 0b1
		req, err := clench.NewHSRequest(protocol, r.DestAddr.Port, []byte(r.DestAddr.FQDN))
		if err != nil {
			return nil, err
		}
		return req.Marshal(), err
	} else {
		req, err := clench.NewHSRequest(protocol, r.DestAddr.Port, []byte(r.DestAddr.IP))
		if err != nil {
			return nil, err
		}
		return req.Marshal(), err
	}
}

func S5ResConvert(reader io.Reader) ([]byte, error) {
	r, err := clench.ReadHSResponse(reader)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(r.STATUS, []byte{0, 0}) {
		log.Printf("ERR: %v's %x ", r, r.STATUS)
	}

	log.Println("channal build")

	return []byte{
		0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10,
	}, nil
}
