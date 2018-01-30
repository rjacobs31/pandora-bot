package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"

	"github.com/rjacobs31/pandora-bot/chat"
	"github.com/rjacobs31/pandora-bot/database"
)

// Config represents the collected settings for the app.
type Config struct {
	Chat     chat.Config
	Database database.Config
}

// LoadConfig attempts to load configuration from all supported
// sources.
//
// Precedence for these sources (from low to high) is:
// - Config in user's home directory
// - Config in the working directory
// - Flags passed to the program
func LoadConfig(config *Config) error {
	if config == nil {
		return errors.New("Invalid config pointer passed to LoadConfig()")
	}

	if LoadFileConfig(configDirName+".toml", config) != nil {
		if defaultDir, err := GetDefaultConfigDir(); err == nil {
			LoadFileConfig(defaultDir, config)
		}
	}

	return LoadFlags(config)
}

var configDirName = "pandora-bot"

// GetDefaultConfigDir attempts to determine the default directory
// in which config is stored.
//
// Unix systems will attempt to use the XDG specification for config
// directories, while other OSes will simply check for a directory
// in `$HOME`.
func GetDefaultConfigDir() (string, error) {
	var configDirLocation string

	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "linux":
		// Use the XDG_CONFIG_HOME variable if it is set, otherwise
		// $HOME/.config/example
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome != "" {
			configDirLocation = xdgConfigHome
		} else {
			configDirLocation = filepath.Join(homeDir, ".config", configDirName)
		}

	default:
		// On other platforms we just use $HOME/.example
		hiddenConfigDirName := "." + configDirName
		configDirLocation = filepath.Join(homeDir, hiddenConfigDirName)
	}

	return configDirLocation, nil
}

// LoadFileConfig attempts to load config from a TOML file.
func LoadFileConfig(configFile string, config *Config) error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return errors.New("Config file does not exist.")
	} else if err != nil {
		return err
	}

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return err
	}

	return nil
}

// LoadFlags attempts to set config using flags set on the command
// line.
func LoadFlags(config *Config) error {
	if config == nil {
		return errors.New("Invalid config pointer passed to LoadConfig()")
	}

	flag.StringVar(&config.Chat.Token, "token", config.Chat.Token, "Discord bot token")
	flag.Parse()
	return nil
}
