package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func init() {
	discord.AddHandler(onGuildUpdate)
	discord.AddHandler(onGuildRemove)
	discord.AddHandler(onChannelRemove)
}

// onGuildUpdate is responsible for ensuring the current permission state is up to date with all voice states.
// It is called during bot startup & after executing linking commands
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
			log.Println("Channel exists in config, but not in state.")
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

// onGuildRemove is responsible for maintaining our config state if the bot is removed from a guild
func onGuildRemove(_ *discordgo.Session, event *discordgo.GuildDelete) {
	configMutex.Lock()
	defer configMutex.Unlock()

	_, guildKnown := config.Guilds[event.ID]
	if !guildKnown {
		return
	}

	delete(config.Guilds, event.ID)
	go saveConfig()
}

// onChannelRemove is responsible for maintaining our config state if one of the linked channel is deleted.
func onChannelRemove(discord *discordgo.Session, event *discordgo.ChannelDelete) {
	// Get a write lock
	configMutex.Lock()
	defer configMutex.Unlock()

	// Check if we know this guild in our config
	channels, guildKnown := config.Guilds[event.GuildID]
	if !guildKnown {
		return
	}

	updated := false

	// If the channel ID matches any of the ones we know, remove the link
	for voice, text := range channels {
		if event.ID == voice || event.ID == text {
			delete(channels, voice)
			updated = true
		}
	}

	if updated {
		// If that was the last channel in this guild, delete the guild too.
		if len(channels) == 0 {
			delete(config.Guilds, event.GuildID)
		}

		go saveConfig()

		// And trigger a guild update if needed
		guild, err := getGuild(discord, event.GuildID)
		if err != nil {
			log.Println("Couldn't fetch guild.", err)
			return
		}

		go onGuildUpdate(discord, &discordgo.GuildCreate{Guild: guild})
	}
}
