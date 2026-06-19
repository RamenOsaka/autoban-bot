package main

import (
	"github.com/bwmarrin/discordgo"
)

func bannedRoleLog(globalName string, userID string, bannedRole string) discordgo.MessageSend {
	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "🔨 Automatic Ban",
				Color:       0xff1100, // Green
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User",
						Value:  globalName + " (`" + userID + "`)",
						Inline: true,
					},
					{
						Name:   "Trigger Role",
						Value:  bannedRole,
						Inline: true,
					},
					{
						Name:   "Reason",
						Value:  "User used a banned role. (role: `" + bannedRole + "`)",
						Inline: false,
					},
				}, 
			},
		},
	}
}