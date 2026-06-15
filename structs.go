package main

import "github.com/bwmarrin/discordgo"

type ServerConfig struct {
	BannedRoles []string
}
type Command struct {
	Definition *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}