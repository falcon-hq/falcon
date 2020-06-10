package cmd

import (
	"fmt"
	remotecmd "github.com/maoxs2/falcon/cmd/remote"
	"github.com/spf13/viper"
	"os"

	localcmd "github.com/maoxs2/falcon/cmd/local"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "falcon",
	Short: "Create secure tunnel for data stream",
	Long:  `Falcon is a proxy tool which helps transfer anything in simple and secure, which can be seen as an alternative of shadowsocks or v2ray. It's for study only.`,
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file, default is config.(yaml, json, toml)")

	rootCmd.AddCommand(localcmd.LocalCmd, remotecmd.RemoteCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.Getwd()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
