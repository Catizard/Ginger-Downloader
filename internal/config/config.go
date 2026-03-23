// Package config manages the local config
package config

import (
	"sync"

	"github.com/spf13/viper"
)

const VERSION = "PREVIEW"

var lock sync.Mutex

func init() {
	viper.SetConfigName("gdownloader")
	viper.SetConfigType("json")
	viper.AddConfigPath("./")
	viper.SafeWriteConfig()
}

type Config struct {
	Initialized       int32
	ClientType        ClientType // LR2 or Raja
	LocalDBPath       string     // Beatoraja's songdata.db or LR2's song.db
	DownloadDirectory string     // Where we download to
}

type ClientType string

var (
	CLIENT_LR2       ClientType = "LR2"
	CLIENT_BEATORAJA ClientType = "Beatoraja"
)

func ReadConfig() (*Config, error) {
	lock.Lock()
	defer lock.Unlock()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{
		Initialized:       viper.GetInt32("Initialized"),
		ClientType:        ClientType(viper.GetString("ClientType")),
		LocalDBPath:       viper.GetString("LocalDBPath"),
		DownloadDirectory: viper.GetString("DownloadDirectory"),
	}

	return config, nil
}

func (conf *Config) WriteConfig() error {
	lock.Lock()
	defer lock.Unlock()
	viper.Set("Initialized", conf.Initialized)
	viper.Set("ClientType", conf.ClientType)
	viper.Set("LocalDBPath", conf.LocalDBPath)
	viper.Set("DownloadDirectory", conf.DownloadDirectory)
	return viper.WriteConfig()
}
