package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Convenience type used to identify discord/twitter IDs
type snowflake = string

var discord *discordgo.Session

func init() {
	if os.Getenv("TOKEN") == "" {
		log.Fatal("Please provide a Discord bot token through the \"TOKEN\" environment variable.")
	}

	// Initialize the bot session
	var err error
	discord, err = discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Open the websocket connection
	if err := discord.Open(); err != nil {
		log.Fatal(err)
	}
	defer discord.Close()

	// Wait for application exit from an OS signal (Ctrl+C for example)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
