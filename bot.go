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

	log.Println("Initializing bot...")
	// Initialize the bot session
	var err error
	discord, err = discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer log.Println("Successfully disconnected.")

	// Open the websocket connection
	if err := discord.Open(); err != nil {
		log.Fatal(err)
	}
	defer discord.Close()

	log.Println("Bot has successfully connected to Discord, now accepting events...")
	log.Println("Use Ctrl+C to shut the bot down.")
	defer log.Println("Shutting down bot...")

	// Wait for application exit from an OS signal (Ctrl+C for example)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
