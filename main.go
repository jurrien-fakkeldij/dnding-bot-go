package main

import (
	"jurrien/dnding-bot/commands"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

var logger *log.Logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
	TimeFormat:      time.DateTime,
})

func main() {
	logger.Info("Starting server!")
	token := os.Getenv("DISCORD_TOKEN")
	StartServer(token)
}

func StartServer(token string) {
	session, err := SetupDiscordBot(token)
	if err != nil {
		logger.Fatal("Something went wrong with setting up the discord bot", "err", err)
	}

	if err = StartDiscordBot(session); err != nil {
		logger.Fatal("Something went wrong opening the discord session", "err", err)
	}

	if err = AddApplicationCommands(session); err != nil {
		logger.Fatal("Something went wrong adding commands for the discord bot", "err", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	logger.Print("Press Ctrl+C to exit")
	<-stop

	if err = RemoveApplicationCommands(session); err != nil {
		logger.Error("Error removing application commands", "err", err)
	}

	if err = StopDiscordBot(session); err != nil {
		logger.Fatal("Could not stop discord session gracefully: %v", err)
	}

	log.Info("Gracefully shutting down.")
}

func SetupDiscordBot(token string) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Error("Invalid bot parameters", "err", err)
		return nil, err
	}
	logger.Info("Started the go discord bot!")
	return session, nil
}

func AddApplicationCommands(session *discordgo.Session) error {
	err := commands.AddPlayerCommands(session)
	if err != nil {
		logger.Error("Error setting up player commands", "err", err)
		return err
	}

	return nil
}

func RemoveApplicationCommands(session *discordgo.Session) error {
	err := commands.RemovePlayerCommands(session)
	if err != nil {
		logger.Error("Error removing player commands", "err", err)
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
