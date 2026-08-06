package main

import "github.com/bwmarrin/discordgo"

type ServerConfig struct {
	BannedRoleID   string
	LogChannelID   string
	DisableAutoBan bool
}
type Command struct {
	Definition *discordgo.ApplicationCommand
	Handler    func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
