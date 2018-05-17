package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func init() {
	discord.AddHandler(onAFK)
}

func onAFK(discord *discordgo.Session, voiceState *discordgo.VoiceStateUpdate) {
	if !voiceState.Deaf && !voiceState.SelfDeaf {
		return
	}

	if voiceState.ChannelID == "" {
		return
	}

	guild, err := getGuild(discord, voiceState.GuildID)
	if err != nil {
		log.Println("Could not get guild info", err)
		return
	}

	if voiceState.ChannelID == guild.AfkChannelID {
		return
	}

	discord.GuildMemberMove(voiceState.GuildID, voiceState.UserID, guild.AfkChannelID)
}
