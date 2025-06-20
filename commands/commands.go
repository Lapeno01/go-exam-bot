package commands

import (
	"fmt"
	"strings"
	"time"

	"exam_bot/database"
	"exam_bot/logger"
	"github.com/bwmarrin/discordgo"
)

func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := strings.ToLower(m.Content)
	args := strings.Fields(content)

	if len(args) == 0 || !strings.HasPrefix(content, "!") {
		return
	}

	command := args[0][1:]
	logger.Info().
		Str("command", command).
		Str("user", m.Author.Username).
		Msg("Processing command")

	switch command {
	case "addexam":
		if len(args) < 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: !addexam <exam_name> <dd.mm.yyyy>")
			if err != nil {
				logger.Error().Err(err).Str("command", command).Msg("Failed to send usage message")
			}
			return
		}
		addExam(s, m, args[1], args[2])
	case "timeleft":
		if len(args) < 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: !timeleft <exam_name>")
			if err != nil {
				logger.Error().Err(err).Str("command", command).Msg("Failed to send usage message")
			}
			return
		}
		timeLeft(s, m, args[1])
	case "deleteexam":
		if len(args) < 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: !deleteexam <exam_name>")
			if err != nil {
				logger.Error().Err(err).Str("command", command).Msg("Failed to send usage message")
			}
			return
		}
		deleteExam(s, m, args[1])
	case "updateexam":
		if len(args) < 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: !updateexam <exam_name> <dd.mm.yyyy>")
			if err != nil {
				logger.Error().Err(err).Str("command", command).Msg("Failed to send usage message")
			}
			return
		}
		updateExam(s, m, args[1], args[2])
	case "listexams":
		listExams(s, m)
	default:
		logger.Warn().Str("command", command).Msg("Unknown command")
	}
}

func addExam(s *discordgo.Session, m *discordgo.MessageCreate, name, dateStr string) {
	err := database.AddExam(name, dateStr)
	if err != nil {
		if strings.Contains(err.Error(), "parse") {
			logger.Error().Err(err).Str("date", dateStr).Msg("Invalid date format")
			_, err := s.ChannelMessageSend(m.ChannelID, "Invalid date format. Use dd.mm.yyyy")
			if err != nil {
				logger.Error().Err(err).Str("command", "addexam").Msg("Failed to send error message")
			}
			return
		}

		if strings.Contains(err.Error(), "past") {
			logger.Warn().Str("date", dateStr).Msg("Attempted to add exam in the past")
			_, err := s.ChannelMessageSend(m.ChannelID, "Cannot add exam in the past")
			if err != nil {
				logger.Error().Err(err).Str("command", "addexam").Msg("Failed to send error message")
			}
			return
		}

		if strings.Contains(err.Error(), "exists") {
			logger.Warn().Str("exam", name).Msg("Exam already exists")
			_, err := s.ChannelMessageSend(m.ChannelID, "Exam with this name already exists")
			if err != nil {
				logger.Error().Err(err).Str("command", "addexam").Msg("Failed to send error message")
			}
			return
		}

		logger.Error().Err(err).Msg("Failed to add exam")
		_, err = s.ChannelMessageSend(m.ChannelID, "Error adding exam")
		if err != nil {
			logger.Error().Err(err).Str("command", "addexam").Msg("Failed to send error message")
		}
		return
	}

	date, _ := time.Parse("02.01.2006", dateStr)
	logger.Info().Str("exam", name).Str("date", date.Format("02.01.2006")).Msg("Exam added")
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Exam %s added for %s", name, date.Format("02.01.2006")))
	if err != nil {
		logger.Error().Err(err).Str("command", "addexam").Msg("Failed to send success message")
	}
}

func timeLeft(s *discordgo.Session, m *discordgo.MessageCreate, name string) {
	exam, err := database.GetExam(name)
	if err != nil {
		logger.Warn().Str("exam", name).Msg("Exam not found")
		_, err := s.ChannelMessageSend(m.ChannelID, "Exam not found")
		if err != nil {
			logger.Error().Err(err).Str("command", "timeleft").Msg("Failed to send error message")
		}
		return
	}

	// Load Germany timezone
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load Europe/Berlin timezone")
		if time.Now().UTC().IsDST() {
			loc = time.FixedZone("CEST", 2*60*60) // UTC+2 (Summer)
		} else {
			loc = time.FixedZone("CET", 1*60*60) // UTC+1 (Winter)
		}
	}

	now := time.Now().In(loc)
	examDate := exam.Date.In(loc)
	duration := examDate.Sub(now)
	if duration < 0 {
		logger.Info().Str("exam", name).Msg("Exam has passed")
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Exam %s has already passed", name))
		if err != nil {
			logger.Error().Err(err).Str("command", "timeleft").Msg("Failed to send error message")
		}
		return
	}

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	logger.Info().Str("exam", name).Int("days", days).Int("hours", hours).Int("minutes", minutes).Msg("Calculated time left")
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
		"Time until **%s**: **%d** days, **%d** hours, **%d** minutes",
		name, days, hours, minutes))
	if err != nil {
		logger.Error().Err(err).Str("command", "timeleft").Msg("Failed to send success message")
	}
}

func deleteExam(s *discordgo.Session, m *discordgo.MessageCreate, name string) {
	err := database.DeleteExam(name)
	if err != nil {
		logger.Warn().Str("exam", name).Msg("Exam not found for deletion")
		_, err := s.ChannelMessageSend(m.ChannelID, "Exam not found")
		if err != nil {
			logger.Error().Err(err).Str("command", "deleteexam").Msg("Failed to send error message")
		}
		return
	}

	logger.Info().Str("exam", name).Msg("Exam deleted")
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Exam %s deleted", name))
	if err != nil {
		logger.Error().Err(err).Str("command", "deleteexam").Msg("Failed to send success message")
	}
}

func updateExam(s *discordgo.Session, m *discordgo.MessageCreate, name, dateStr string) {
	err := database.UpdateExam(name, dateStr)
	if err != nil {
		if strings.Contains(err.Error(), "parse") {
			logger.Error().Err(err).Str("date", dateStr).Msg("Invalid date format")
			_, err := s.ChannelMessageSend(m.ChannelID, "Invalid date format. Use dd.mm.yyyy")
			if err != nil {
				logger.Error().Err(err).Str("command", "updateexam").Msg("Failed to send error message")
			}
			return
		}

		if strings.Contains(err.Error(), "past") {
			logger.Warn().Str("date", dateStr).Msg("Attempted to update exam to past date")
			_, err := s.ChannelMessageSend(m.ChannelID, "Cannot update exam to a past date")
			if err != nil {
				logger.Error().Err(err).Str("command", "updateexam").Msg("Failed to send error message")
			}
			return
		}

		logger.Error().Err(err).Msg("Failed to update exam")
		_, err = s.ChannelMessageSend(m.ChannelID, "Exam not found")
		if err != nil {
			logger.Error().Err(err).Str("command", "updateexam").Msg("Failed to send error message")
		}
		return
	}

	date, _ := time.Parse("02.01.2006", dateStr)
	logger.Info().Str("exam", name).Str("date", date.Format("02.01.2006")).Msg("Exam updated")

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Exam %s updated to %s", name, date.Format("02.01.2006")))
	if err != nil {
		logger.Error().Err(err).Str("command", "updateexam").Msg("Failed to send success message")
	}
}

func listExams(s *discordgo.Session, m *discordgo.MessageCreate) {
	exams, err := database.ListExams()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list exams")
		_, err := s.ChannelMessageSend(m.ChannelID, "Error listing exams")
		if err != nil {
			logger.Error().Err(err).Str("command", "listexams").Msg("Failed to send error message")
		}
		return
	}

	if len(exams) == 0 {
		logger.Info().Msg("No exams scheduled")
		_, err := s.ChannelMessageSend(m.ChannelID, "No exams scheduled")
		if err != nil {
			logger.Error().Err(err).Str("command", "listexams").Msg("Failed to send message")
		}
		return
	}

	var response strings.Builder
	response.WriteString("Scheduled exams:\n")
	for _, exam := range exams {
		response.WriteString(fmt.Sprintf("- %s: %s\n", exam.Name, exam.Date.Format("02.01.2006")))
	}

	logger.Info().Msg("Listed exams")
	_, err = s.ChannelMessageSend(m.ChannelID, response.String())

	if err != nil {
		logger.Error().Err(err).Str("command", "listexams").Msg("Failed to send success message")
	}
}
