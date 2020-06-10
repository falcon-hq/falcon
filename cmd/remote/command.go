package remotecmd

import (
	logging "github.com/ipfs/go-log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	servePort    int
	key          string
	chunkSize    int
	enableSnappy bool
	logLevel     string
)

var RemoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Server daemon for remote device",
	Run:   run,
}

func init() {
	RemoteCmd.PersistentFlags().IntVarP(&servePort, "serve-port", "l", 18000, "Source directory to read from")
	RemoteCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "key for data encryption")
	RemoteCmd.PersistentFlags().IntVar(&chunkSize, "chunk-size", 2<<10, "chunk size")
	RemoteCmd.PersistentFlags().BoolVarP(&enableSnappy, "enable-snappy", "z", false, "enable snappy")
	RemoteCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level")

	viper.BindPFlag("serve-port", RemoteCmd.PersistentFlags().Lookup("serve-port"))
	viper.BindPFlag("key", RemoteCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("chunk-size", RemoteCmd.PersistentFlags().Lookup("chunk-size"))
	viper.BindPFlag("enable-snappy", RemoteCmd.PersistentFlags().Lookup("enable-snappy"))
	viper.BindPFlag("log-level", RemoteCmd.PersistentFlags().Lookup("log-level"))
}

func run(cmd *cobra.Command, args []string) {

	var remote *Remote
	if len(viper.ConfigFileUsed()) == 0 {
		lvl, err := logging.LevelFromString(logLevel)
		if err != nil {
			panic(err)
		}
		logging.SetAllLoggers(lvl)

		remote = NewRemote(servePort, key, chunkSize, enableSnappy)
	} else {
		lvl, err := logging.LevelFromString(viper.GetString("log-level"))
		if err != nil {
			panic(err)
		}
		logging.SetAllLoggers(lvl)

		remote = NewRemote(viper.GetInt("serve-port"), viper.GetString("key"), viper.GetInt("chunk-size"), viper.GetBool("enable-snappy"))
	}

	go remote.MainLoop()

	select {}
}
