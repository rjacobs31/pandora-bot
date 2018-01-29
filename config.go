package main

import (
	_ "github.com/BurntSushi/toml"

	"github.com/rjacobs31/pandora-bot/chat"
	"github.com/rjacobs31/pandora-bot/database"
)

type Config struct {
	Chat     chat.Config
	Database database.Config
}
