package main

import (
	"exam_bot/commands"
	"exam_bot/config"
	"exam_bot/database"
	"exam_bot/logger"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Load configuration first to get LogPath
	cfg, err := config.Load("prod.yaml")
	if err != nil {
		_, _ = os.Stderr.WriteString("Failed to load configuration: " + err.Error() + "\n")
		return
	}
	logger.Info().Str("file", "prod.yaml").Msg("Configuration loaded successfully")

	// Initialize logger with LogPath from config
	err = logger.Init(cfg.LogPath)
	if err != nil {
		_, _ = os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		return
	}
	logger.Info().Str("log_path", cfg.LogPath).Msg("Logger initialized successfully")
	logger.Info().Msg("Starting Discord bot")

	// Initialize database
	err = database.Init(cfg.StoragePath)
	if err != nil {
		logger.Error().Err(err).Str("storage_path", cfg.StoragePath).Msg("Failed to initialize database")
		return
	}
	defer database.Close()
	logger.Info().Str("storage_path", cfg.StoragePath).Msg("Database initialized successfully")

	// Create Discord session
	dg, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		logger.Error().Err(err).Int("token_length", len(cfg.BotToken)).Msg("Failed to create Discord session")
		return
	}
	logger.Info().Int("token_length", len(cfg.BotToken)).Msg("Discord session created successfully")

	// Register command handler
	dg.AddHandler(commands.HandleMessageCreate)
	logger.Info().Msg("Command handler registered successfully")

	// Open Discord connection
	err = dg.Open()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to open Discord connection")
		return
	}
	logger.Info().Str("bot_user_id", dg.State.User.ID).Msg("Discord connection opened successfully")

	// Wait for termination signal
	logger.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Shutdown
	logger.Info().Msg("Received shutdown signal, shutting down bot")
	if err := dg.Close(); err != nil {
		logger.Error().Err(err).Msg("Failed to close Discord session")
	}
	logger.Info().Msg("Bot shutdown completed")
}
