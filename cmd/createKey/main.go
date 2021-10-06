package main

import (
	"io/ioutil"
	"os"

	"github.com/IpsoVeritas/crypto"
	"github.com/IpsoVeritas/logger"
	"github.com/IpsoVeritas/realm/pkg/version"
	"github.com/spf13/viper"
)

func main() {
	viper.AutomaticEnv()
	viper.SetDefault("log_formatter", "text")
	viper.SetDefault("log_level", "debug")
	viper.SetDefault("key", "./realm.pem")

	logger.SetOutput(os.Stdout)
	logger.SetFormatter(viper.GetString("log_formatter"))
	logger.SetLevel(viper.GetString("log_level"))
	logger.AddContext("service", "createKey")
	logger.AddContext("version", version.Version)

	fn := viper.GetString("key")
	_, err := os.Stat(fn)
	if err != nil {
		logger.Infof("Creating key %s", fn)

		key, err := crypto.NewKey()
		if err != nil {
			logger.Fatal(err)
		}

		kb, err := crypto.MarshalToPEM(key)
		if err != nil {
			logger.Fatal(err)
		}

		if err := ioutil.WriteFile(viper.GetString("key"), kb, 0600); err != nil {
			logger.Fatal(err)
		}
	} else {
		logger.Infof("Key %s already exists", fn)
	}
}
