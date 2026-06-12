package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	// Setting up discord token
	discordToken := loadToken()
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	// Adding handlers
	dg.AddHandler(ready)
	dg.AddHandler(handlerName)

	// Setting up discord Intents
	dg.Identify.Intents = discordgo.IntentGuildMembers

	err = dg.Open()
	if err != nil {
		log.Println("Error opening Discord session: ", err)
	}

	log.Println("Autoban is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateGameStatus(0, "Dev Env")
}

func handlerName(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	fmt.Print("Username :", m.User.Username, " | Roles : ")
	roleList, err := s.GuildRoles(m.GuildID)
	if err != nil {
		log.Println("Error trying to read the Guild's roles: ", err)
	}

	for i := 0; i < len(m.Roles); i++ {
		for j := 0; j < len(roleList); j++ {
			if roleList[j].ID == m.Roles[i] {
				fmt.Print(roleList[j].Name, " ")
			}
		}
	}
	fmt.Println()
}

func loadToken() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token, exists := os.LookupEnv("DISCORD_TOKEN")
	if !exists {
		log.Println("DISCORD_TOKEN is not set!")
	}

	return token
}
