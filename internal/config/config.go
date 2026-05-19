// Package config manages the local config
package config

import (
	"errors"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/Catizard/bmsdb"
	"github.com/rotisserie/eris"
	"muzzammil.xyz/jsonc"
)

const (
	VERSION = "PREVIEW"
)

var configFilePath = "./config.jsonc"

// MoveConfigFilePath changes the config file path,
// presented for only internal test usage
func MoveConfigFilePath(fp string) {
	configFilePath = fp
}

var (
	lock sync.Mutex
	// Snapshot is a ref for other modules to read the configurations without
	// triggering the real disk load. During runtime there's no write operation
	// will every happen so this is a safe strategy to use
	Snapshot atomic.Pointer[Config]
)

type ClientType string

var (
	CLIENT_LR2       ClientType = "LR2"
	CLIENT_BEATORAJA ClientType = "Beatoraja"
)

type Config struct {
	ClientType           ClientType // LR2 or Raja
	GameInstallationPath string
	DownloadDirectory    string // Where we download to
}

// InitConfig initialize the config module. If the config file hasn't been created or
// it fails to be parsed, an error would be returned. Otherwise it returns the config
// instance.
//
// Possible erros: ErrorNoConfig and ErrorInvalidConfig.
func InitConfig() (*Config, error) {
	if _, err := os.Stat(configFilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrorNoConfig
		}
	}
	conf, err := ReadConfig()
	if err != nil {
		return nil, err
	}

	if err := validateConfig(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// ReadConfig reads the config file on disk and refresh the snapshot
func ReadConfig() (*Config, error) {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Open(configFilePath)
	if err != nil {
		return nil, eris.Wrap(err, "read config: open file")
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, eris.Wrap(err, "read config: read file")
	}
	var conf Config
	data := jsonc.ToJSON(b)
	log.Printf("data: %s", string(data))
	if err := jsonc.Unmarshal(b, &conf); err != nil {
		return nil, eris.Wrap(err, "read config: unmarshal")
	}
	Snapshot.Store(&conf)
	return &conf, nil
}

// validateConfig validates the config passed in. Errors will be grouped
// into an ErrorInvalidConfig
func validateConfig(conf *Config) error {
	var errValidateConfig ErrorInvalidConfig
	if conf.ClientType == "" {
		errValidateConfig.Append(eris.Errorf("No bms client type provided. Expecting LR2 or Beatoraja (case sensitive)"))
	} else {
		if conf.ClientType != CLIENT_LR2 && conf.ClientType != CLIENT_BEATORAJA {
			errValidateConfig.Append(eris.Errorf("Wrong client type: %s. Expecting LR2 or Beatoraja (case sensitive)", conf.ClientType))
		}
	}

	if conf.GameInstallationPath == "" {
		errValidateConfig.Append(eris.Errorf("No game installation path provided"))
	} else {
		if err := isDirectory(conf.GameInstallationPath); err != nil {
			errValidateConfig.Append(err)
		} else {
			switch conf.ClientType {
			case CLIENT_LR2:
				if err := bmsdb.ValidateLR2Installation(conf.GameInstallationPath); err != nil {
					errValidateConfig.Append(eris.Errorf("%s is not a valid LR2 installation directory: %v", conf.GameInstallationPath, err))
				}
			case CLIENT_BEATORAJA:
				if err := bmsdb.ValidateBeatorajaInstallation(conf.GameInstallationPath); err != nil {
					errValidateConfig.Append(eris.Errorf("%s is not a valid Beatoraja installation directory: %v", conf.GameInstallationPath, err))
				}
			}
		}
	}

	if conf.DownloadDirectory == "" {
		errValidateConfig.Append(eris.Errorf("No download directory provided"))
	} else {
		if err := isDirectory(conf.DownloadDirectory); err != nil {
			errValidateConfig.Append(err)
		}
	}

	if len(errValidateConfig.Errors) == 0 {
		return nil
	}
	return errValidateConfig
}

// isDirectory returns nil if dir is a directory
func isDirectory(dir string) error {
	if stat, err := os.Stat(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return eris.Errorf("%s is not exist", dir)
		}
		return eris.Errorf("%s is unable to stat", dir)
	} else if !stat.IsDir() {
		return eris.Errorf("%s is not a directory", dir)
	}
	return nil
}
