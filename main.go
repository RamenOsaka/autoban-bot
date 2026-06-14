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
var serverID = "1260943648695255140"
var commandList = []discordgo.ApplicationCommand{}

func main() {
	// Setting up discord token
	discordToken, appID := loadEnv()
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	// Adding handlers
	dg.AddHandler(ready)
	dg.AddHandler(handlerRoleUpdate)
	dg.AddHandler(handlerRoleOnJoin)
	dg.AddHandler(handlerInteraction)

	// Setting up discord Intents
	dg.Identify.Intents = discordgo.IntentGuildMembers

	err = dg.Open()
	if err != nil {
		log.Println("Error opening Discord session: ", err)
	}

	addRole := &discordgo.ApplicationCommand{
		Name:        "addrole",
		Description: "Adds a role to the autoban role list",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "role to be added",
				Required:    true,
			},
		},
	}
	commandList = append(commandList, *addRole)

	pingCommand := &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Hello there",
	}
	commandList = append(commandList, *pingCommand)

	for i := 0; i < len(commandList); i++ {
		_, err = dg.ApplicationCommandCreate(appID, serverID, &commandList[i])
		if err != nil {
			log.Println("Error creating a new command : ", err)
		}
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

func handlerInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		if i.ApplicationCommandData().Name == "ping" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
		} else if i.ApplicationCommandData().Name == "addrole" {
			if slices.Contains(bannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, serverID).ID) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, serverID).Name + "** is already in the banned list role.",
					},
				})
			} else {
				bannedRoles = append(bannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, serverID).ID)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, serverID).Name + "** has been added to the banned list role.",
					},
				})
			}
		}
	}
}

func loadEnv() (string, string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token, exists := os.LookupEnv("DISCORD_TOKEN")
	if !exists {
		log.Println("DISCORD_TOKEN is not set!")
	}
	appID, exists := os.LookupEnv("APP_ID")
	if !exists {
		log.Println("APP_ID is not set!")
	}
	return token, appID
}
