package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const configFileName = "config.json"

// guildLinks is a map used for one specific guild, the key is the voice channel, the value is the linked text channel
type guildChannels = map[snowflake]snowflake

// channelList is the global registry of guilds that we have voice-text channel links for
type channelList = map[snowflake]guildChannels

var (
	configMutex sync.RWMutex
	config      = struct {
		// Guilds contains all voice-text-channel links per guild.
		// The key is the voice channel ID, the value is the text channel ID.
		Guilds channelList `json:"guilds"`
	}{
		Guilds: make(channelList),
	}
)

func init() {
	// If the config file doesn't exist, create it.
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		if err = saveConfig(); err != nil {
			log.Fatal(err)
		}

		return
	}

	// Or else open it
	f, err := os.Open(configFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// And read it in JSON format to our config struct
	if err = json.NewDecoder(f).Decode(&config); err != nil {
		log.Fatal(err)
	}
}

func saveConfig() error {
	configMutex.RLock()
	defer configMutex.RUnlock()

	f, err := os.OpenFile(configFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	e := json.NewEncoder(f)
	e.SetIndent("", "    ")

	return e.Encode(config)
}
