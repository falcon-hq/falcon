package main

import (
	"crypto/aes"
	"crypto/cipher"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/maoxs2/falcon/common"
	aeadconn "github.com/maoxs2/go-aead-conn"
	"github.com/maoxs2/go-socks5"
)

var listenPort = flag.Int("l", 18000, "listen port")
var key = flag.String("k", "", "key for encryption")
var chunkSize = flag.Int("chunk", 2<<(10+4), "chunk size, default 16k")

func main() {
	flag.Parse()

	serveAddr := fmt.Sprintf("0.0.0.0:%d", *listenPort)
	l, err := net.Listen("tcp", serveAddr)
	if err != nil {
		panic(err)
	}

	log.Printf("remote server is serving on %s, key: %s", serveAddr, *key)
	fmt.Printf("You can run client with the following arguments: ./client -k %s -r [Remote Server IP]:%d -chunk %d \n", *key, *listenPort, *chunkSize)
	block, err := aes.NewCipher(common.Sha3Sum256([]byte(*key)))
	if err != nil {
		panic(err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	conf := &socks5.Config{}
	s5server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	server := &server{
		Server: s5server,
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("accepted conn from %s", conn.RemoteAddr())
		go handleConn(conn, aead, server)
	}
}

func handleConn(conn net.Conn, aead cipher.AEAD, s5 *server) {
	cryptoConn := aeadconn.NewAEADCompressConn(common.GetSeed(), *chunkSize, conn, aead)

	err := s5.handleConn(cryptoConn)
	if err != nil {
		log.Println(err)
	}
}
