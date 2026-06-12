package main

import (
	"log"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var bannedRoles = []string{"1514834166892597420"}

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
	dg.AddHandler(handlerRoleUpdate)
	dg.AddHandler(handlerRoleOnJoin)

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

func banRole(s *discordgo.Session, m *discordgo.Member) {
	for i := 0; i < len(m.Roles); i++ {
		if slices.Contains(bannedRoles, m.Roles[i]) {
			s.GuildBanCreateWithReason(m.GuildID, m.User.ID, "Picked a banned role", 7)
		}
	}
}

func handlerRoleOnJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	banRole(s, m.Member)
}

func handlerRoleUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	banRole(s, m.Member)
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
