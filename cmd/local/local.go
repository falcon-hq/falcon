package localcmd

import (
	"crypto/aes"
	"crypto/cipher"
	"log"
	"net"

	"github.com/maoxs2/falcon/common"
	"github.com/maoxs2/falcon/local"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var localSocks5Addr string
var remoteAddr string
var key string
var chunkSize int
var authUser string
var authPass string
var enableSnappy bool

var LocalCmd = &cobra.Command{
	Use:   "local",
	Short: "Client daemon for local device",
	Run:   run,
}

func init() {
	LocalCmd.Flags().StringVarP(&localSocks5Addr, "listen-socks5", "l", "127.0.0.1:10008", "Source directory to read from")
	LocalCmd.Flags().StringVarP(&remoteAddr, "remote-addr", "r", "", "remote addr")
	LocalCmd.Flags().StringVarP(&key, "key", "k", "", "key for data encryption")
	LocalCmd.Flags().IntVar(&chunkSize, "chunk-size", 2<<10, "chunk size")
	LocalCmd.Flags().StringVarP(&authUser, "auth-username", "u", "", "username value for socks authorization")
	LocalCmd.Flags().StringVarP(&authPass, "auth-password", "p", "", "password value for socks authorization")
	LocalCmd.Flags().BoolP("enable-snappy", "z", false, "enable snappy")

	LocalCmd.MarkFlagRequired("remote-addr")

	viper.BindPFlag("listen-socks5", LocalCmd.Flags().Lookup("listen-socks5"))
	viper.BindPFlag("remote-addr", LocalCmd.Flags().Lookup("remote-addr"))
	viper.BindPFlag("key", LocalCmd.Flags().Lookup("key"))
	viper.BindPFlag("chunk-size", LocalCmd.Flags().Lookup("chunk-size"))
	viper.BindPFlag("auth-username", LocalCmd.Flags().Lookup("auth-username"))
	viper.BindPFlag("auth-password", LocalCmd.Flags().Lookup("auth-password"))
	viper.BindPFlag("enable-snappy", LocalCmd.Flags().Lookup("enable-snappy"))
}

func run(cmd *cobra.Command, args []string) {
	l, err := net.Listen("tcp", localSocks5Addr)
	if err != nil {
		panic(err)
	}

	log.Printf("serving on local: %s", localSocks5Addr)
	log.Printf("connect to %s with key %s", remoteAddr, key)
	block, err := aes.NewCipher(common.Sha3Sum256([]byte(key)))
	if err != nil {
		panic(err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	local := local.NewLocal(remoteAddr, chunkSize, authUser, authPass, enableSnappy)

	go func() {
		for {
			localConn, err := l.Accept()
			if err != nil {
				log.Println(err)
				continue
			}

			if err := local.HandleAuth(localConn); err != nil {
				log.Println(err)
			}

			log.Println("New App Conn")
			go local.HandleConn(localConn, aead)
		}
	}()

	for {
		select {}
	}
}
