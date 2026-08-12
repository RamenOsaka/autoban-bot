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
				Description: "**" + username + " has been banned**",
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

func errorLog(err error) discordgo.MessageSend { 
	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "⚠️ An **error** occured!",
				Color: 0xb5af3a,
				Footer: &discordgo.MessageEmbedFooter{
					Text: time.Now().Format("2006-01-02 15:04:05"),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
			},
		},
	}
}