package main

import (
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

var commands = map[string]Command{
	"ping": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "ping",
			Description:              "Hello there",
			DefaultMemberPermissions: &defaultPerms,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
		},
		Handler: handlePing,
	},
	"setrole": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "setrole",
			Description:              "sets the role to automatically ban",
			DefaultMemberPermissions: &defaultPerms,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "role to be banned upon picking",
					Required:    true,
				},
			},
		},
		Handler: handleSetRole,
	},
	"disableautoban": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "disableautoban",
			Description:              "disables the automatic ban",
			DefaultMemberPermissions: &defaultPerms,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "bool",
					Description: "set to true to disable automatic ban",
					Required:    true,
				},
			},
		},
		Handler: handleDisableAutoban,
	},
	"deletedata": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "deletedata",
			Description:              "deletes all of the server's settings. Also happens automatically when the bot leaves the server.",
			DefaultMemberPermissions: &defaultPerms,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
		},
		Handler: handleDeleteData,
	},
	"showbannedrole": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "showbannedrole",
			Description:              "Shows the banned role",
			DefaultMemberPermissions: &defaultPerms,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
		},
		Handler: handleShowBannedRole,
	},
	"setlogchannel": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "setlogchannel",
			Description:              "Select channel to output logs to.",
			DefaultMemberPermissions: &defaultPerms,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "log channel",
					Required:    true,
				},
			},
		},
		Handler: handleSetLogChannel,
	},
}

func handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})
}

func handleDeleteData(s *discordgo.Session, i *discordgo.InteractionCreate) {
	delete(serverConfigs, i.GuildID)
	saveConfig()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "All data has been deleted!",
		},
	})
}

func handleShowBannedRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	name := ""
	listContent := "Banned role : **"
	roleList, err := s.GuildRoles(i.GuildID)
	if err != nil {
		log.Println("Error trying to retrieve role list :", err)
	}

	if len(serverConfigs[i.GuildID].BannedRoleID) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "The banned role hasn't been set",
			},
		})
		return
	} else {
		for _, role := range roleList {
			if role.ID == serverConfigs[i.GuildID].BannedRoleID {
				name = role.Name
				break
			}
		}
	}

	listContent = listContent + name + "**"
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: listContent,
		},
	})
}

func handleSetRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if serverConfigs[i.GuildID].BannedRoleID == i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** is already set as the banned role.",
			},
		})
	} else {
		config := serverConfigs[i.GuildID]
		config.BannedRoleID = i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID
		serverConfigs[i.GuildID] = config
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** has been set as the banned role.",
			},
		})
	}
	saveConfig()
}

func handleDisableAutoban(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := serverConfigs[i.GuildID]
	config.DisableAutoBan = i.ApplicationCommandData().Options[0].BoolValue()
	serverConfigs[i.GuildID] = config
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Automatic banning is not set to **" + strconv.FormatBool(i.ApplicationCommandData().Options[0].BoolValue()) + "**.",
		},
	})
	saveConfig()
}

func handleSetLogChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := serverConfigs[i.GuildID]
	config.LogChannelID = i.ApplicationCommandData().Options[0].ChannelValue(s).ID
	serverConfigs[i.GuildID] = config
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Channel **" + i.ApplicationCommandData().Options[0].ChannelValue(s).Name + "** has been set as the log channel.",
		},
	})
	saveConfig()
}
