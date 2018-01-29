package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rjacobs31/pandora-bot/chat"
)

func main() {
	config := Config{}
	flag.StringVar(&config.Chat.Token, "token", "", "Discord bot token")
	flag.Parse()

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
