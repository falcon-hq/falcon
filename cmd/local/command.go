package localcmd

import (
	logging "github.com/ipfs/go-log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	socks5Addr   string
	targetAddr   string
	key          string
	chunkSize    int
	authUser     string
	authPass     string
	enableSnappy bool
	logLevel     string
)

var LocalCmd = &cobra.Command{
	Use:   "local",
	Short: "Client daemon for local device",
	Run:   run,
}

func init() {
	LocalCmd.PersistentFlags().StringVarP(&socks5Addr, "socks5-addr", "l", "127.0.0.1:10008", "Source directory to read from")
	LocalCmd.PersistentFlags().StringVarP(&targetAddr, "target-addr", "r", "", "target remote addr")
	LocalCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "key for data encryption")
	LocalCmd.PersistentFlags().IntVar(&chunkSize, "chunk-size", 2<<10, "chunk size")
	LocalCmd.PersistentFlags().StringVarP(&authUser, "auth-username", "u", "", "username value for socks authorization")
	LocalCmd.PersistentFlags().StringVarP(&authPass, "auth-password", "p", "", "password value for socks authorization")
	LocalCmd.PersistentFlags().BoolVarP(&enableSnappy, "enable-snappy", "z", false, "enable snappy")
	LocalCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level")

	viper.BindPFlag("socks5-addr", LocalCmd.PersistentFlags().Lookup("socks5-addr"))
	viper.BindPFlag("target-addr", LocalCmd.PersistentFlags().Lookup("target-addr"))
	viper.BindPFlag("key", LocalCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("chunk-size", LocalCmd.PersistentFlags().Lookup("chunk-size"))
	viper.BindPFlag("auth-username", LocalCmd.PersistentFlags().Lookup("auth-username"))
	viper.BindPFlag("auth-password", LocalCmd.PersistentFlags().Lookup("auth-password"))
	viper.BindPFlag("enable-snappy", LocalCmd.PersistentFlags().Lookup("enable-snappy"))
	viper.BindPFlag("log-level", LocalCmd.PersistentFlags().Lookup("log-level"))
}

func run(cmd *cobra.Command, args []string) {
	var local *Local
	if len(viper.ConfigFileUsed()) == 0 {
		lvl, err := logging.LevelFromString(logLevel)
		if err != nil {
			panic(err)
		}
		logging.SetAllLoggers(lvl)

		local = NewLocal(socks5Addr, targetAddr, key, chunkSize, authUser, authPass, enableSnappy)
	} else {
		lvl, err := logging.LevelFromString(viper.GetString("log-level"))
		if err != nil {
			panic(err)
		}
		logging.SetAllLoggers(lvl)

		local = NewLocal(
			viper.GetString("socks5-addr"), viper.GetString("target-addr"),
			viper.GetString("key"), viper.GetInt("chunk-size"), viper.GetString("auth-username"),
			viper.GetString("auth-password"), viper.GetBool("enable-snappy"))
	}

	go local.MainLoop()

	select {}
}
