package main

import (
	"github.com/bwmarrin/discordgo"
	"time"
)

func bannedRoleLog(username string, userID string, bannedRole string) discordgo.MessageSend {
	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "🔨 Automatic Ban",
				Color: 0xff1100,
				Footer: &discordgo.MessageEmbedFooter{
					Text: time.Now().Format("2006-01-02 15:04:05"),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User",
						Value:  username + " (`" + userID + "`)",
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
