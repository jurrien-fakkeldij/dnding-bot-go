package main

import (
	"context"
	"fmt"
	"jurrien/dnding-bot/commands"
	"jurrien/dnding-bot/database"
	"os"
	"os/signal"
	"syscall"
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
	databaseDSN := os.Getenv("DB_DSN")
	ctx := context.Background()
	StartServer(ctx, token, databaseDSN)
}

func StartServer(ctx context.Context, token string, dbDSN string) {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := &database.Config{
		DSN: dbDSN,
	}

	logger.Info("Setting up database", "dsn", dbDSN)
	db, err := database.SetupDB(ctx, config)

	if err != nil {
		logger.Fatal("failed to setup database", "err", err)
	}

	logger.Info("Setup discord bot", "token", token)
	session, err := SetupDiscordBot(token)

	if err != nil {
		logger.Fatal("Something went wrong with setting up the discord bot", "err", err)
	}

	if err = StartDiscordBot(session); err != nil {
		logger.Fatal("Something went wrong opening the discord session", "err", err)
	}

	if err = AddApplicationCommands(session, db, logger); err != nil {
		logger.Fatal("Something went wrong adding commands for the discord bot", "err", err)
	}

	logger.Print("Press Ctrl+C to exit")
	<-ctx.Done()

	_, cancel = context.WithTimeout(context.Background(), 29*time.Second)
	defer cancel()

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

func AddApplicationCommands(session *discordgo.Session, database *database.DB, logger *log.Logger) error {
	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		commandName := interaction.ApplicationCommandData().Name
		if h, ok := commands.AllCommandHandlers[commandName]; ok {
			err := h(session, database, logger, interaction)
			if err != nil {
				logger.Error("Error in a command", "command", commandName, "error", err)
			}
		}
	})

	for index, command := range commands.AllCommands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, "", command)
		if err != nil {
			return fmt.Errorf("Cannot create '%v' command: %v", command.Name, err)
		}
		commands.RegisteredCommands[index] = cmd
	}

	return nil
}

func RemoveApplicationCommands(session *discordgo.Session) error {
	for _, command := range commands.RegisteredCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, "", command.ID)
		if err != nil {
			return fmt.Errorf("Cannot delete '%v' command: %v", command.Name, err)
		}
	}
	return nil
}

func StartDiscordBot(session *discordgo.Session) error {
	return session.Open()
}

func StopDiscordBot(session *discordgo.Session) error {
	return session.Close()
}
