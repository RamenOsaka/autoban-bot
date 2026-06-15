package main

import (
	"log"
	"slices"
	"strings"

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
	names := []string{}
	listContent := "Banned roles : "
	roleList, err := s.GuildRoles(i.GuildID)
	if err != nil {
		log.Println("Error trying to retrieve role list :", err)
	}

	if len(serverConfigs[i.GuildID].BannedRoles) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "The banned role list is empty",
			},
		})
		return
	} else {
		for _, roleID := range serverConfigs[i.GuildID].BannedRoles {
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
	if slices.Contains(serverConfigs[i.GuildID].BannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** is already in the banned list role.",
			},
		})
	} else {
		config := serverConfigs[i.GuildID]
		config.BannedRoles = append(serverConfigs[i.GuildID].BannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID)
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

func handlerRemoveRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if slices.Contains(serverConfigs[i.GuildID].BannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID) {
		config := serverConfigs[i.GuildID]
		config.BannedRoles = slices.Delete(config.BannedRoles, slices.Index(config.BannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID), slices.Index(config.BannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID)+1)
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
