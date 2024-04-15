package main

import (
	"jurrien/dnding-bot/commands"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

func main() {
	log.Println("Starting server!")
	token := os.Getenv("DISCORD_TOKEN")
	StartServer(token)
}

func StartServer(token string) {
	session, err := SetupDiscordBot(token)
	if err != nil {
		log.Fatalf("Something went wrong with setting up the discord bot: %v", err)
	}

	if err = StartDiscordBot(session); err != nil {
		log.Fatalf("Something went wrong opening the discord session: %v", err)
	}

	if err = AddApplicationCommands(session); err != nil {
		log.Fatalf("Something went wrong adding commands for the discord bot: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if err = RemoveApplicationCommands(session); err != nil {
		log.Printf("Error removing application commands: %v", err)
	}

	if err = StopDiscordBot(session); err != nil {
		log.Fatalf("Could not stop discord session gracefully: %v", err)
	}

	log.Println("Gracefully shutting down.")
}

func SetupDiscordBot(token string) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Printf("Invalid bot parameters: %v", err)
		return nil, err
	}
	log.Println("Started the go discord bot!")
	return session, nil
}

func AddApplicationCommands(session *discordgo.Session) error {
	err := commands.AddPlayerCommands(session)
	if err != nil {
		log.Printf("Error setting up player commands: %v", err)
		return err
	}

	return nil
}

func RemoveApplicationCommands(session *discordgo.Session) error {
	err := commands.RemovePlayerCommands(session)
	if err != nil {
		log.Printf("Error removing player commands: %v", err)
		return err
	}
	return nil
}

func StartDiscordBot(session *discordgo.Session) error {
	return session.Open()
}

func StopDiscordBot(session *discordgo.Session) error {
	return session.Close()
}
