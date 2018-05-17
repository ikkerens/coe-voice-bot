package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	discord.AddHandler(onCommandEvent)
}

func onCommandEvent(discord *discordgo.Session, event *discordgo.MessageCreate) {
	if !strings.HasPrefix(event.Content, "!") {
		return
	}

	// Check if the user has the MANAGE_CHANNELS permission
	serverPerms, _ := getPermissionsFromMessage(discord, event)
	if serverPerms&discordgo.PermissionManageChannels != discordgo.PermissionManageChannels {
		// Ignore
		return
	}

	args := strings.Split(event.Content, " ")

	switch strings.ToLower(args[0]) {
	case "!voicelink":
		linkCommand(discord, event, args[1:])
	case "!voiceunlink":
		unlinkCommand(discord, event, args[1:])
	case "!voicelinklist":
		list(discord, event)
	}

	// Silently fail if there's an unknown command
}

func linkCommand(discord *discordgo.Session, event *discordgo.MessageCreate, args []string) {
	// Check if the command was invoked correctly
	if len(args) != 2 {
		discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" Usage of this command:\n"+
			"```\n"+
			"!voicelink <voiceChannelID> <textChannelID|textChannelMention>\n"+
			"```")
		return
	}

	// Get the voice channel instance
	voice, err := getChannel(discord, args[0])
	if err != nil {
		discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" I'm sorry, I could not find that voice channel.")
		return
	}

	// Get the text channel instance
	text, err := getChannel(discord, strings.Trim(args[1], "<#>"))
	if err != nil {
		discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" I'm sorry, I could not find that text channel.")
		return
	}

	// Ensure they're of the same type
	if voice.Type != discordgo.ChannelTypeGuildVoice || text.Type != discordgo.ChannelTypeGuildText {
		discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" The first argument needs to be a voice channel, "+
			"the second argument needs to be a text channel.")
		return
	}

	// Get the channel the command was invoked in
	channel, err := getChannel(discord, event.ChannelID)
	if err != nil {
		log.Println("Could not fetch channel from despite us being able to earlier")
		return
	}

	// Make sure it was invoked in the correct guild
	if channel.GuildID != voice.GuildID || channel.GuildID != text.GuildID {
		discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" The channels provided both need"+
			"to be in the same server as where you execute the command.")
	}

	// Add it to the list
	configMutex.Lock()
	list, exists := config.Guilds[channel.GuildID]
	if !exists {
		list = make(guildChannels)
		config.Guilds[channel.GuildID] = list
	}
	list[voice.ID] = text.ID
	configMutex.Unlock()
	go saveConfig()

	// Send a confirmation
	discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" Success! I've linked the voice channel "+
		voice.Name+" to the text channel "+text.Mention()+".")

	// And trigger a guild update
	guild, err := getGuild(discord, channel.GuildID)
	if err != nil {
		log.Println("Couldn't fetch guild.", err)
		return
	}
	go onGuildUpdate(discord, &discordgo.GuildCreate{Guild: guild})
}

func unlinkCommand(discord *discordgo.Session, event *discordgo.MessageCreate, args []string) {
	// Check if the command was invoked correctly
	if len(args) != 1 {
		discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" Usage of this command:\n"+
			"```\n"+
			"!voiceunlink <voiceChannelID>\n"+
			"```")
		return
	}

	// Get the channel the command was invoked in
	channel, err := getChannel(discord, event.ChannelID)
	if err != nil {
		log.Println("Could not fetch channel from despite us being able to earlier")
		return
	}

	configMutex.Lock()
	defer configMutex.Unlock()

	// Check if this guild even has any registered channels
	channels, guildKnown := config.Guilds[channel.GuildID]
	if !guildKnown {
		discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" I know no registered channels for this server.")
		return
	}

	// Check if the requested channel is registered
	_, channelRegistered := channels[args[0]]
	if !channelRegistered {
		discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" That is not a registered voice channel in this server.")
		return
	}

	// Remove it from the list
	delete(channels, args[0])
	if len(channels) == 0 {
		delete(config.Guilds, channel.GuildID)
	}
	go saveConfig()

	// Send a confirmation
	discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" Success! I've unlinked that voice channel!")

	// And trigger a guild update
	guild, err := getGuild(discord, channel.GuildID)
	if err != nil {
		log.Println("Couldn't fetch guild.", err)
		return
	}

	go onGuildUpdate(discord, &discordgo.GuildCreate{Guild: guild})
}

func list(discord *discordgo.Session, event *discordgo.MessageCreate) {
	// Get the channel the command was invoked in
	channel, err := getChannel(discord, event.ChannelID)
	if err != nil {
		log.Println("Could not fetch channel from despite us being able to earlier")
		return
	}

	configMutex.RLock()
	defer configMutex.RUnlock()

	channels, guildKnown := config.Guilds[channel.GuildID]
	if !guildKnown {
		discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" I know no registered channels for this server.")
		return
	}

	description := event.Author.Mention() + "These are the voice channels I have currently linked to text channels:\n"
	found := false
	for voiceID, textID := range channels {
		voice, err := getChannel(discord, voiceID)
		if err != nil {
			continue // Ignore, invalid link
		}

		text, err := getChannel(discord, textID)
		if err != nil {
			continue // Ignore, invalid link
		}

		found = true
		description += fmt.Sprintf("\nThe voice channel \"%s\" (%s) is linked to %s (%s).", voice.Name, voice.ID, text.Mention(), text.ID)
	}

	if !found {
		discord.ChannelMessageSend(event.ChannelID, event.Author.Mention()+" I know no registered channels for this server.")
		return
	}

	discord.ChannelMessageSend(event.ChannelID, description)
}
