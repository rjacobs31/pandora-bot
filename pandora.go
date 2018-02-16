package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rjacobs31/pandora-bot/chat"
	"github.com/rjacobs31/pandora-bot/chat/handlers"
	"github.com/rjacobs31/pandora-bot/database"
)

func main() {
	config := Config{}
	if err := LoadConfig(&config); err != nil {
		log.Fatalln("Could not load config: %s", err)
		return
	}

	if config.Chat.Token == "" {
		log.Fatal("Error: Needs bot token.")
		return
	}

	log.Println("Starting Discord client")
	client, err := chat.New(config.Chat.Token)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Adding message handlers")
	InitHandlers(client, config)

	log.Println("Opening Discord connection")
	client.Start()

	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(
		sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		os.Interrupt,
		os.Kill,
	)
	log.Println("Bot running. Press Ctrl+C to exit.")
	<-sc
}

func InitHandlers(c *chat.ChatClient, config Config) {
	c.AddHandler(new(handlers.PingHandler))
	c.AddHandler(new(handlers.BananaLoveHandler))

	db, err := database.InitialiseDB(config.Database)
	if err != nil {
		return
	}

	c.AddHandler(handlers.NewFactoidRegisterHandler(&db.FactoidManager))
	c.AddHandler(handlers.NewRetortHandler(&db.FactoidManager))
}
