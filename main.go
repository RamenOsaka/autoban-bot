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

var serverConfigs = map[string]ServerConfig{}
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
    "ping":    handlePing,
    "addrole": handleAddRole,
}

//remove in prod
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

	// Opening websocket
	err = dg.Open()
	if err != nil {
		log.Println("Error opening Discord session: ", err)
	}

	// Adding commands
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
		if slices.Contains(serverConfigs[m.GuildID].bannedRoles, m.Roles[i]) {
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
	if _, exists := serverConfigs[i.GuildID]; !exists {
		serverConfigs[i.GuildID] = ServerConfig{}
	}

	if i.Type == discordgo.InteractionApplicationCommand {
		if handler, exists := commandHandlers[i.ApplicationCommandData().Name]; exists {
			handler(s, i)
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
	} else if token == "" {
		log.Println("DISCORD_TOKEN is empty!")
	}
	appID, exists := os.LookupEnv("APP_ID")
	if !exists {
		log.Println("APP_ID is not set!")
	} else if token == "" {
		log.Println("APP_ID is empty!")
	}
	return token, appID
}

// Slash command functions

func handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
}

func handleAddRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if slices.Contains(serverConfigs[i.GuildID].bannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** is already in the banned list role.",
					},
				})
			} else {
				config := serverConfigs[i.GuildID]
				config.bannedRoles = append(serverConfigs[i.GuildID].bannedRoles, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID)
				serverConfigs[i.GuildID] = config
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "Role **" + i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).Name + "** has been added to the banned list role.",
					},
				})
			}
}