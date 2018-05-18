package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func init() {
	discord.AddHandler(onVoiceStateUpdate)
}

func onVoiceStateUpdate(discord *discordgo.Session, voiceState *discordgo.VoiceStateUpdate) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	guild, exists := config.Guilds[voiceState.GuildID]
	if !exists {
		return
	}

	// First, find the channels the user already has overwrites for
	toRemove := make(map[*discordgo.Channel]bool)
	for voiceID, textID := range guild {
		channel, err := getChannel(discord, textID)
		if err != nil {
			log.Println("could not check channel overwrites", err)
			continue
		}

		overwrite := getOverwriteByID(channel, voiceState.UserID, "member")
		if overwrite == nil {
			continue // No overwrites in place, don't need to remove
		}

		if voiceID == voiceState.ChannelID && !voiceState.Deaf && !voiceState.SelfDeaf {
			toRemove[channel] = false // We're about to grant access to this channel, so no need to remove it
			continue
		}

		_, exists := toRemove[channel]
		if !exists {
			toRemove[channel] = true
		}
	}

	// After finding them, actually remove them
	for text, remove := range toRemove {
		if remove {
			log.Printf("Removing override for user %s in channel #%s.\n", getUserName(discord, voiceState.GuildID, voiceState.UserID), text.Name)
			if err := discord.ChannelPermissionDelete(text.ID, voiceState.UserID); err != nil {
				log.Println("Could not remove override.", err)
			}
		}
	}

	// Second, if the user joins a channel, give them access to the linked channel, if registered
	if voiceState.ChannelID == "" {
		return // Didn't join
	}

	if voiceState.Deaf || voiceState.SelfDeaf {
		return // Don't show channel if the user is deafened.
	}

	textID, exists := guild[voiceState.ChannelID]
	if !exists {
		return // Channel not linked
	}

	// Check if the overwrite already exists
	text, err := getChannel(discord, textID)
	if err != nil {
		log.Println("Could not fetch voice-chat channel", err)
		return
	}

	// Only set read permissions if this member doesn't already have an overwrite
	overwrite := getOverwriteByID(text, voiceState.UserID, "member")
	if overwrite == nil {
		log.Printf("Creating override for user %s in channel #%s.\n", getUserName(discord, voiceState.GuildID, voiceState.UserID), text.Name)
		if err = discord.ChannelPermissionSet(textID, voiceState.UserID, "member", discordgo.PermissionReadMessages, 0); err != nil {
			log.Println("Could not create channel override.", err)
		}
	}
}
