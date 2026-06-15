package main

import (
	"slices"

	"github.com/bwmarrin/discordgo"
)

func handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})
}

func handlerListBannedRoles(s *discordgo.Session, i *discordgo.InteractionCreate) {
	listContent := "Banned roles :"
	// for _, roleID := range serverConfigs[i.GuildID].bannedRoles {
	// 	for roles := range s.GuildRoles(i.GuildID) {

	// 		listContent = listContent + " **" + +"**"
	// 	}
	// }
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: listContent,
		},
	})
}

func handleAddRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if slices.Contains(serverConfigs[i.GuildID].bannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** is already in the banned list role.",
			},
		})
	} else {
		config := serverConfigs[i.GuildID]
		config.bannedRoles = append(serverConfigs[i.GuildID].bannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID)
		serverConfigs[i.GuildID] = config
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** has been added to the banned role list.",
			},
		})
	}
}

func handlerRemoveRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if slices.Contains(serverConfigs[i.GuildID].bannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID) {
		config := serverConfigs[i.GuildID]
		config.bannedRoles = slices.Delete(config.bannedRoles, slices.Index(config.bannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID), slices.Index(config.bannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID)+1)
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
}
