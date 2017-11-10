package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/namsral/flag"

	pandora "github.com/rjacobs31/pandora-bot"
	"github.com/rjacobs31/pandora-bot/bolt"
)

// Variables used for command line parameters
var (
	Token  string
	DBPath string

	Client pandora.DataClient
)

func init() {
	flag.String(flag.DefaultConfigFlagname, "", "path to config file")
	flag.StringVar(&Token, "pandora_token", "", "Bot Token")
	flag.StringVar(&DBPath, "pandora_db", "pandora.boltdb", "BoltDB database file path")
	flag.Parse()
}

func main() {
	fmt.Println("Application starting")
	c := bolt.NewClient(DBPath)
	if err := c.Open(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("BoltDB client connected")
	defer c.Close()
	Client = c

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	fmt.Println("Discord session created")

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	fmt.Println("Discord handlers registered")

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("Discord connected")
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
