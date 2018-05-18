package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func init() {
	discord.AddHandler(onAFK)
}

// onAFK is responsible for moving users that deafen themselves to the Guilds AFK channel
func onAFK(discord *discordgo.Session, voiceState *discordgo.VoiceStateUpdate) {
	// This event handler should only return if the user is deafened
	if !voiceState.Deaf && !voiceState.SelfDeaf {
		return
	}

	// No need to do anything if the user is leaving a channel
	if voiceState.ChannelID == "" {
		return
	}

	// Get the guild information, so we can see the AFK channel
	guild, err := getGuild(discord, voiceState.GuildID)
	if err != nil {
		log.Println("Could not get guild info", err)
		return
	}

	// No need to move if the user is already in the AFK channel
	if voiceState.ChannelID == guild.AfkChannelID {
		return
	}

	// Move the user
	log.Printf("Moving user %s to the guild AFK channel because they are deafened.\n", getUserName(discord, voiceState.GuildID, voiceState.UserID))
	if err = discord.GuildMemberMove(voiceState.GuildID, voiceState.UserID, guild.AfkChannelID); err != nil {
		log.Println("Could not move member to AFK channel", err)
	}
}
