package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func init() {
	discord.AddHandler(onGuildUpdate)
}

func onGuildUpdate(discord *discordgo.Session, newGuild *discordgo.GuildCreate) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	// Check if this guild is a registered guild
	guild, exists := config.Guilds[newGuild.ID]
	if !exists {
		return
	}

	seen := make(map[string]bool)

	// First, add existing voice states.
	for _, state := range newGuild.VoiceStates {
		seen[state.UserID] = true
		go onVoiceStateUpdate(discord, &discordgo.VoiceStateUpdate{VoiceState: state})
	}

	// Second, remove existing overwrites that are no longer valid
	for _, textChannelID := range guild {
		textChannel, err := getChannel(discord, textChannelID)
		if err != nil {
			log.Println("Channel exists in GuildCreate, but not in state.")
			continue
		}

		for _, overwrite := range textChannel.PermissionOverwrites {
			if overwrite.Type != "member" {
				continue
			}

			_, isInVoice := seen[overwrite.ID]
			if isInVoice {
				continue
			}

			// We have an overwrite for this user on a registered channel, they're not in the seen map, so remove them.
			go onVoiceStateUpdate(discord, &discordgo.VoiceStateUpdate{VoiceState: &discordgo.VoiceState{
				GuildID: newGuild.ID,
				UserID:  overwrite.ID,
			}})
		}
	}
}
