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

var serverConfigs = map[string]ServerConfig{}

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
	dg.AddHandler(handlerGuildDelete)

	// Setting up discord Intents
	dg.Identify.Intents = discordgo.IntentGuildMembers

	// Opening websocket
	err = dg.Open()
	if err != nil {
		log.Println("Error opening Discord session: ", err)
	}

	// Creating commands
	loadCommands(dg, appID, testGuildID)

	log.Println("Autoban is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func banRole(s *discordgo.Session, m *discordgo.Member) string {
	roleName := ""
	if slices.Contains(m.Roles, serverConfigs[m.GuildID].BannedRoleID) && !serverConfigs[m.GuildID].DisableAutoBan {
		s.GuildBanCreateWithReason(m.GuildID, m.User.ID, "Picked a banned role", 7)

		serverRoleList, err := s.GuildRoles(m.GuildID)
		if err != nil {
			log.Println("Couldn't get the list of the server's role : ", err)
		}

		index := slices.IndexFunc(serverRoleList, func(u *discordgo.Role) bool {
			return u.ID == serverConfigs[m.GuildID].BannedRoleID
		})
		roleName = serverRoleList[index].Name
	}
	return roleName
}

func loadCommands(s *discordgo.Session, appID string, guildID string) {
	_, err := s.ApplicationCommandBulkOverwrite(appID, "", []*discordgo.ApplicationCommand{})
	if err != nil {
		log.Fatal("Could not delete the global commands: ", err)
	}
	_, err = s.ApplicationCommandBulkOverwrite(appID, guildID, []*discordgo.ApplicationCommand{})
	if err != nil {
		log.Fatal("Could not delete the server commands commands: ", err)
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

func saveConfig() {
	config, err := json.Marshal(serverConfigs)
	if err != nil {
		log.Println("Could not save guild data: ", err)
	}
	os.WriteFile(configFilePath, config, 0644)
}

func loadConfig() {
	var data map[string]ServerConfig
	config, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Println(configFilePath, " Hasn't been created yet : ", err)
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
		log.Fatal("DISCORD_TOKEN is not set!")
	} else if token == "" {
		log.Fatal("DISCORD_TOKEN is empty!")
	}
	appID, exists := os.LookupEnv("APP_ID")
	if !exists {
		log.Fatal("APP_ID is not set!")
	} else if appID == "" {
		log.Fatal("APP_ID is empty!")
	}
	return token, appID
}
