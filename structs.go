package main

import "github.com/bwmarrin/discordgo"

type ServerConfig struct {
	BannedRolesID []string
	LogChannelID string
}
type Command struct {
	Definition *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}