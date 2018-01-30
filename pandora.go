package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rjacobs31/pandora-bot/chat"
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
	client.Start()

	log.Println("Opening Discord connection")
	err = client.Open()
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
