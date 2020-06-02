package remotecmd

import (
	"crypto/aes"
	"crypto/cipher"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/maoxs2/falcon/common"
	"github.com/maoxs2/falcon/remote"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listenPort int
var key string
var chunkSize int
var enableSnappy bool

var RemoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Server daemon for remote device",
	Run:   run,
}

func init() {
	RemoteCmd.Flags().IntVarP(&listenPort, "listen-port", "l", 18000, "Source directory to read from")
	RemoteCmd.Flags().StringVarP(&key, "key", "k", "", "key for data encryption")
	RemoteCmd.Flags().IntVar(&chunkSize, "chunk-size", 2<<10, "chunk size")
	RemoteCmd.Flags().BoolP("enable-snappy", "z", false, "enable snappy")

	viper.BindPFlag("listen-port", RemoteCmd.Flags().Lookup("listen-socks5"))
	viper.BindPFlag("key", RemoteCmd.Flags().Lookup("key"))
	viper.BindPFlag("chunk-size", RemoteCmd.Flags().Lookup("chunk-size"))
	viper.BindPFlag("enable-snappy", RemoteCmd.Flags().Lookup("enable-snappy"))
}

func run(cmd *cobra.Command, args []string) {
	flag.Parse()

	serveAddr := fmt.Sprintf("0.0.0.0:%d", listenPort)
	l, err := net.Listen("tcp", serveAddr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("remote server is serving on %s, key: %s \n", serveAddr, key)
	fmt.Printf("You can run falcon with the following arguments in your local device: \n ./falcon local -k %s -r [Remote Server IP]:%d -chunk %d \n", key, listenPort, chunkSize)
	block, err := aes.NewCipher(common.Sha3Sum256([]byte(key)))
	if err != nil {
		panic(err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	remote := remote.NewRemote(chunkSize, enableSnappy)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("accepted conn from %s", conn.RemoteAddr())
		go remote.HandleConn(conn, aead)
	}
}
