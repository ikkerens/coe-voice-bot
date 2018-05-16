package main

/*
All functions in this file are copied from https://github.com/ikkerens/gophbot.
Implementation based on https://discordapp.com/developers/docs/topics/permissions
Ported by Rens "Ikkerens" Rikkerink
*/

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"
)

// getGuild attempts to get a guild instance from the shard state cache, and if none exists, attempts to obtain it
// from the Discord API. Will err if the guild does not exist or if this guild is unreachable for the bot.
func getGuild(discord *discordgo.Session, guildID snowflake) (*discordgo.Guild, error) {
	guild, err := discord.State.Guild(guildID)
	if err == nil {
		return guild, nil
	}

	guild, err = discord.Guild(guildID)
	if err != nil {
		return nil, err
	}
	discord.State.GuildAdd(guild)

	return guild, nil
}

// getChannel attempts to get a channel instance from the shard state cache, and if none exists, attempts to obtain it
// from the Discord API. Will err if the channel does not exist or the bot is not in the guild it belongs to.
func getChannel(discord *discordgo.Session, channelID snowflake) (*discordgo.Channel, error) {
	channel, err := discord.State.Channel(channelID)
	if err == nil {
		return channel, nil
	}

	channel, err = discord.Channel(channelID)
	if err != nil {
		return nil, err
	}
	discord.State.ChannelAdd(channel)

	return channel, err
}

// getGuildMember will attempt to obtain a member instance for this guild member.
// Will err if this user is not a member of this guild
func getGuildMember(discord *discordgo.Session, guildID, userID snowflake) (*discordgo.Member, error) {
	member, err := discord.State.Member(guildID, userID)
	if err == nil {
		return member, nil
	}

	member, err = discord.GuildMember(guildID, userID)
	if err != nil {
		return nil, err
	}
	discord.State.MemberAdd(member)

	return member, nil
}

// GetRole will attempt to obtain a role instance for this role
// Will err if the given ID is not a role in the guild.
func GetRole(discord *discordgo.Session, guildID, roleID snowflake) (*discordgo.Role, error) {
	role, err := discord.State.Role(guildID, roleID)
	if err == nil {
		return role, nil
	}

	roles, err := discord.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		if role.ID == roleID {
			discord.State.RoleAdd(guildID, role)
			return role, nil
		}
	}

	return nil, errors.New("role does not exist or is not part of that guild")
}

// getOverwrite will attempt to obtain the PermissionOverwrite instance for the target (Role, Member or User)
// Will return nil if no such overwrite exists
func getOverwrite(channel *discordgo.Channel, target interface{}) *discordgo.PermissionOverwrite {
	var id snowflake
	var typ string

	switch t := target.(type) {
	case *discordgo.Role:
		id = t.ID
		typ = "role"
	case *discordgo.Member:
		id = t.User.ID
		typ = "user"
	case *discordgo.User:
		id = t.ID
		typ = "user"
	default:
		panic(errors.New("invalid overwrite target type"))
	}

	return getOverwriteByID(channel, id, typ)
}

// getOverwriteByID will attempt to obtain the PermissionOverwrite instance for the given ID and type ("role" or "user")
// Will return nil if no such overwrite exists
func getOverwriteByID(channel *discordgo.Channel, id snowflake, typ string) *discordgo.PermissionOverwrite {
	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.Type == typ && overwrite.ID == id {
			return overwrite
		}
	}

	return nil
}

// computeBasePermissions calculates the permissions a guild member has, outside the scope of a channel
func computeBasePermissions(discord *discordgo.Session, member *discordgo.Member, guild *discordgo.Guild) (int, error) {
	if guild.OwnerID == member.User.ID {
		return discordgo.PermissionAll, nil
	}

	everyone, err := GetRole(discord, guild.ID, guild.ID)
	if err != nil {
		return 0, err
	}

	permissions := everyone.Permissions
	for _, roleID := range member.Roles {
		role, err := GetRole(discord, guild.ID, roleID)
		if err != nil {
			return 0, err
		}

		permissions |= role.Permissions
	}

	if permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
		return discordgo.PermissionAll, nil
	}

	return permissions, nil
}

// computeOverwrites calculates the permissions a channel member has, given the guilds base permissions
func computeOverwrites(discord *discordgo.Session, basePermissions int, member *discordgo.Member, channel *discordgo.Channel) (int, error) {
	if basePermissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
		return discordgo.PermissionAll, nil
	}

	everyone, err := GetRole(discord, channel.GuildID, channel.GuildID)
	if err != nil {
		return 0, err
	}

	everyoneOverwrite := getOverwrite(channel, everyone)
	if everyoneOverwrite != nil {
		basePermissions &= ^everyoneOverwrite.Deny
		basePermissions |= everyoneOverwrite.Allow
	}

	var allow, deny int
	for _, roleID := range member.Roles {
		overwrite := getOverwriteByID(channel, roleID, "role")
		if overwrite != nil {
			allow |= overwrite.Allow
			deny |= overwrite.Deny
		}
	}

	basePermissions &= ^deny
	basePermissions |= allow

	memberOverwrite := getOverwriteByID(channel, member.User.ID, "user")
	if memberOverwrite != nil {
		basePermissions &= ^memberOverwrite.Deny
		basePermissions |= memberOverwrite.Allow
	}

	return basePermissions, nil
}

func getPermissionsFromMessage(discord *discordgo.Session, event *discordgo.MessageCreate) (server, channel int) {
	channelI, err := getChannel(discord, event.ChannelID)
	if err != nil {
		log.Println("Could not fetch channel", err)
		return 0, 0
	}

	guild, err := getGuild(discord, channelI.GuildID)
	if err != nil {
		log.Println("Could not fetch guild", err)
		return 0, 0
	}

	member, err := getGuildMember(discord, guild.ID, event.Author.ID)
	if err != nil {
		log.Println("Could not fetch guild member", err)
		return 0, 0
	}

	server, err = computeBasePermissions(discord, member, guild)
	if err != nil {
		log.Println("Could not fetch permissions for server", err)
		return 0, 0
	}

	channel, err = computeOverwrites(discord, server, member, channelI)
	if err != nil {
		log.Println("Could not fetch permissions for channel", err)
		return 0, 0
	}

	return
}
