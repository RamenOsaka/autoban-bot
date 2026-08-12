package main

import (
	"github.com/bwmarrin/discordgo"
)

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "👁️ looking out for bots")
}

func handlerRoleOnJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	bannedRoleSelected, err := banRole(s, m.Member)
	if err != nil {
		data := errorLog(err)
		s.ChannelMessageSendComplex(serverConfigs[m.Member.GuildID].LogChannelID, &data)
	}
	if bannedRoleSelected != "" {
		data := bannedRoleLog(m.User.Username, m.User.ID, bannedRoleSelected)
		s.ChannelMessageSendComplex(serverConfigs[m.Member.GuildID].LogChannelID, &data)
	}
}

func handlerRoleUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	bannedRoleSelected, err := banRole(s, m.Member)
	if err != nil {
		data := errorLog(err)
		s.ChannelMessageSendComplex(serverConfigs[m.Member.GuildID].LogChannelID, &data)
	}
	if bannedRoleSelected != "" {
		data := bannedRoleLog(m.User.GlobalName, m.User.ID, bannedRoleSelected)
		s.ChannelMessageSendComplex(serverConfigs[m.Member.GuildID].LogChannelID, &data)
	}
}

func handlerGuildDelete(s *discordgo.Session, g *discordgo.GuildDelete) {
	delete(serverConfigs, g.ID)
	saveConfig()
}

func handlerInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if _, exists := serverConfigs[i.GuildID]; !exists {
		serverConfigs[i.GuildID] = ServerConfig{}
	}

	if i.Type == discordgo.InteractionApplicationCommand {
		if cmd, exists := commands[i.ApplicationCommandData().Name]; exists {
			cmd.Handler(s, i)
		}
	}
}
