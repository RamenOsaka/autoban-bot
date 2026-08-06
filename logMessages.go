package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func bannedRoleLog(username string, userID string, bannedRole string) discordgo.MessageSend {
	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "🔨 Member banned",
				Description: username + " has been banned",
				Color:       0xff1100,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User",
						Value:  username + " (`" + userID + "`)",
						Inline: true,
					},
					{
						Name:   "Reason",
						Value:  "User picked a banned role. (role: `" + bannedRole + "`)",
						Inline: false,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Autoban",
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
	}
}
