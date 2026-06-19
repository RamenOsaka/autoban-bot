package main

import (
	"log"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
)
var commands = map[string]Command{
	    "ping": {
        Definition: &discordgo.ApplicationCommand{
            Name:        "ping",
            Description: "Hello there",
        },
        Handler: handlePing,
    },
    "addrole": {
        Definition: &discordgo.ApplicationCommand{
            Name:        "addrole",
            Description: "Adds a role to the autoban role list",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionRole,
                    Name:        "role",
                    Description: "role to be added",
                    Required:    true,
                },
            },
        },
        Handler: handleAddRole,
    },
	"removerole": {
        Definition: &discordgo.ApplicationCommand{
            Name:        "removerole",
            Description: "Removes a role to the autoban role list",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionRole,
                    Name:        "role",
                    Description: "role to be removed",
                    Required:    true,
                },
            },
        },
        Handler: handleRemoveRole,
    },
	"listbannedroles": {
        Definition: &discordgo.ApplicationCommand{
            Name:        "listbannedroles",
            Description: "Lists all banned roles.",
        },
        Handler: handleListBannedRoles,
    },
	"setlogchannel": {
		Definition: &discordgo.ApplicationCommand{
			Name: "setlogchannel",
			Description: "Select channel to output logs to.",
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

func handleListBannedRoles(s *discordgo.Session, i *discordgo.InteractionCreate) {
	names := []string{}
	listContent := "Banned roles : "
	roleList, err := s.GuildRoles(i.GuildID)
	if err != nil {
		log.Println("Error trying to retrieve role list :", err)
	}

	if len(serverConfigs[i.GuildID].BannedRolesID) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "The banned role list is empty",
			},
		})
		return
	} else {
		for _, roleID := range serverConfigs[i.GuildID].BannedRolesID {
			for _, role := range roleList {
				if role.ID == roleID  {
					names = append(names, role.Name)
				}
			}
		}
	}

	listContent = listContent + strings.Join(names, ", ")
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: listContent,
		},
	})
}

func handleAddRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if slices.Contains(serverConfigs[i.GuildID].BannedRolesID, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** is already in the banned list role.",
			},
		})
	} else {
		config := serverConfigs[i.GuildID]
		config.BannedRolesID = append(serverConfigs[i.GuildID].BannedRolesID, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID)
		serverConfigs[i.GuildID] = config
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** has been added to the banned role list.",
			},
		})
	}
	saveConfig()
}

func handleRemoveRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if slices.Contains(serverConfigs[i.GuildID].BannedRolesID, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID) {
		config := serverConfigs[i.GuildID]
		config.BannedRolesID = slices.Delete(config.BannedRolesID, slices.Index(config.BannedRolesID, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID), slices.Index(config.BannedRolesID, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID)+1)
		serverConfigs[i.GuildID] = config
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** has been removed from the banned role list.",
			},
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** isn't in the banned role list.",
			},
		})
	}
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