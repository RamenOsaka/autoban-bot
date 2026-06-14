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
						Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** has been added to the banned list role.",
					},
				})
			}
}