package main

import (
	"encoding/base64"
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
	discordToken, appID, tempEncryptionKey := loadEnv()
	encryptionKey = tempEncryptionKey
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
		return
	}
	encrypted, err := encryptData(config)
	if err != nil {
		log.Println("Could not encrypt guild data: ", err)
		return
	}
	os.WriteFile(configFilePath, encrypted, 0644)
}

func loadConfig() {
	var data map[string]ServerConfig
	encrypted, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Println(configFilePath, " Hasn't been created yet : ", err)
		serverConfigs = map[string]ServerConfig{}
		return
	} else if len(encrypted) == 0 {
		serverConfigs = map[string]ServerConfig{}
		return
	}

	config, err := decryptData(encrypted)
	if err != nil {
		log.Fatal("Could not decrypt ", configFilePath, ": ", err)
	}

	json.Unmarshal(config, &data)
	serverConfigs = data
}

func loadEnv() (string, string, []byte) {
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

	base64Key, exists := os.LookupEnv("ENCRYPTION_KEY")
	if !exists || base64Key == "" {
		log.Fatal("ENCRYPTION_KEY is not set! Generate one with: openssl rand -base64 32")
	}
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		log.Fatal("ENCRYPTION_KEY is not valid base64: " + err.Error())
	}
	if len(key) != 32 {
		log.Fatal("ENCRYPTION_KEY must decode to 32 bytes for AES-256")
	}
	return token, appID, key
}
