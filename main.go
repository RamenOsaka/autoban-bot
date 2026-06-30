package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var configFileName = "config.json"
var defaultPerms int64 = discordgo.PermissionAdministrator
var serverConfigs = map[string]ServerConfig{}

// test server for devs
var guildID string = ""

func main() {
	// Setting up discord token
	discordToken, appID := loadEnv()
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	//Loading data
	loadConfig()

	// Adding handlers
	dg.AddHandler(ready)
	dg.AddHandler(handlerRoleUpdate)
	dg.AddHandler(handlerRoleOnJoin)
	dg.AddHandler(handlerInteraction)

	// Setting up discord Intents
	dg.Identify.Intents = discordgo.IntentGuildMembers

	// Opening websocket
	err = dg.Open()
	if err != nil {
		log.Println("Error opening Discord session: ", err)
	}

	// Creating commands
	loadCommands(dg, appID, guildID)

	log.Println("RamenBot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func banRole(s *discordgo.Session, m *discordgo.Member) string {
	roleName := ""
	for i := 0; i < len(m.Roles); i++ {
		if slices.Contains(serverConfigs[m.GuildID].BannedRolesID, m.Roles[i]) {
			s.GuildBanCreateWithReason(m.GuildID, m.User.ID, "Picked a banned role", 7)

			roleList, err := s.GuildRoles(m.GuildID)
			if err != nil {
				log.Println("Couldn't get the list of the server's role : ", err)
			}

			index := slices.IndexFunc(roleList, func(u *discordgo.Role) bool {
				return u.ID == m.Roles[i]
			})
			roleName = roleList[index].Name
		}
	}
	return roleName
}

func saveConfig() {
	config, err := json.Marshal(serverConfigs)
	if err != nil {
		log.Println("Could not transform serverConfigs into json data: ", err)
	}
	os.WriteFile(configFileName, config, 0644)
}

func loadCommands(s *discordgo.Session, appID string, guildID string) {
	_, err := s.ApplicationCommandBulkOverwrite(appID, "", []*discordgo.ApplicationCommand{})
	if err != nil {
		log.Println("Could not delete the global commands: ", err)
	}
	_, err = s.ApplicationCommandBulkOverwrite(appID, guildID, []*discordgo.ApplicationCommand{})
	if err != nil {
		log.Println("Could not delete the server commands commands: ", err)
	}

	var applicationCommandList []*discordgo.ApplicationCommand
	for _, cmd := range commands {
		applicationCommandList = append(applicationCommandList, cmd.Definition)
	}
	_, err = s.ApplicationCommandBulkOverwrite(appID, guildID, applicationCommandList)
	if err != nil {
		log.Println("Couldn't register commands: ", err)
	}
}

func loadConfig() {
	var data map[string]ServerConfig
	config, err := os.ReadFile(configFileName)
	if err != nil {
		log.Println(configFileName+" Hasn't been created yet : ", err)
		serverConfigs = map[string]ServerConfig{}
		return
	} else if len(config) == 0 {
		serverConfigs = map[string]ServerConfig{}
		return
	}

	json.Unmarshal(config, &data)
	serverConfigs = data
}

func loadEnv() (string, string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token, exists := os.LookupEnv("DISCORD_TOKEN")
	if !exists {
		log.Println("DISCORD_TOKEN is not set!")
	} else if token == "" {
		log.Println("DISCORD_TOKEN is empty!")
	}
	appID, exists := os.LookupEnv("APP_ID")
	if !exists {
		log.Println("APP_ID is not set!")
	} else if appID == "" {
		log.Println("APP_ID is empty!")
	}
	return token, appID
}
