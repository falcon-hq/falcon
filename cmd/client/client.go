package main

import (
	"crypto/aes"
	"crypto/cipher"
	"flag"
	"io"
	"log"
	"net"

	"github.com/maoxs2/falcon/common"
	aeadconn "github.com/maoxs2/go-aead-conn"
	"github.com/maoxs2/go-socks5"
)

var localSocks5Addr = flag.String("l", "127.0.0.1:10008", "local addr")
var remoteAddr = flag.String("r", "", "remote addr")
var key = flag.String("k", "", "key for encryption (and auth)")
var chunkSize = flag.Int("chunk", 2<<(10+4), "chunk size")
var authUser = flag.String("u", "", "user value for socks")
var authPass = flag.String("p", "", "pass value for socks")

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", *localSocks5Addr)
	if err != nil {
		panic(err)
	}

	log.Printf("serving on local: %s", *localSocks5Addr)
	log.Printf("connect to %s with key %s", *remoteAddr, *key)
	block, err := aes.NewCipher(common.Sha3Sum256([]byte(*key)))
	if err != nil {
		panic(err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	conf := &socks5.Config{}
	if len(*authUser) > 0 {
		var c socks5.StaticCredentials
		c[*authUser] = *authPass
		conf.AuthMethods = append(conf.AuthMethods, &socks5.UserPassAuthenticator{
			Credentials: c,
		})
	}

	s5server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	s := &server{
		Server: s5server,
	}

	go func() {
		for {
			localConn, err := l.Accept()
			if err != nil {
				log.Println(err)
				continue
			}

			if err := s.handleAuth(localConn); err != nil {
				log.Println(err)
			}

			log.Println("New App Conn")
			go handleConn(localConn, aead)
		}
	}()

	for {
		select {}
	}
}

func handleConn(lc net.Conn, aead cipher.AEAD) {
	externalConn, err := net.Dial("tcp", *remoteAddr)
	if err != nil {
		log.Printf("failed connecting to remote server: %s: %s", externalConn.RemoteAddr(), err)
	}

	cryptoConn := aeadconn.NewAEADCompressConn(common.GetSeed(), *chunkSize, externalConn, aead)

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
