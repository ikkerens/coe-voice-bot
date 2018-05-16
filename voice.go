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
	for _, channelID := range guild {
		if channelID == voiceState.ChannelID && !voiceState.Deaf && !voiceState.SelfDeaf {
			continue // We're about to grant access to this channel, so no need to remove it
		}

		channel, err := getChannel(discord, channelID)
		if err != nil {
			log.Println("could not check channel overwrites", err)
			continue
		}

		overwrite := getOverwriteByID(channel, voiceState.UserID, "member")
		if overwrite == nil {
			continue // No overwrites in place, don't need to remove
		}

		discord.ChannelPermissionDelete(channelID, overwrite.ID)
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
	overwrite := getOverwriteByID(text, voiceState.UserID, "member")

	if overwrite == nil {
		discord.ChannelPermissionSet(textID, voiceState.UserID, "member", discordgo.PermissionReadMessages, 0)
	}
}
