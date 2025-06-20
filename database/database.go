package database

import (
	"database/sql"
	"exam_bot/logger"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var db *sql.DB

type Exam struct {
	Name string
	Date time.Time
}

func Init(storagePath string) error {
	var err error
	db, err = sql.Open("sqlite3", storagePath)
	if err != nil {
		logger.Error().Err(err).Str("storage_path", storagePath).Msg("Failed to open SQLite database")
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS exams (
			name TEXT PRIMARY KEY,
			date TEXT NOT NULL
		)
	`)

	if err != nil {
		logger.Error().Err(err).Str("storage_path", storagePath).Msg("Failed to create exams table")
		return err
	}
	logger.Info().Str("storage_path", storagePath).Msg("Database initialized successfully")
	return nil
}

func Close() {
	if db != nil {
		err := db.Close()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to close database")
			return
		}
		logger.Info().Msg("Database closed successfully")
	}
}

func AddExam(name, dateStr string) error {
	date, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		logger.Error().Err(err).Str("date", dateStr).Msg("Invalid date format")
		return err
	}

	if date.Before(time.Now()) {
		logger.Warn().Str("date", dateStr).Msg("Attempted to add exam in the past")
		return fmt.Errorf("cannot add exam in the past")
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM exams WHERE name = ?", name).Scan(&count)
	if err != nil {
		logger.Error().Err(err).Str("name", name).Msg("Failed to check existing exam")
		return err
	}

	if count > 0 {
		logger.Warn().Str("name", name).Msg("Exam already exists")
		return fmt.Errorf("exam already exists")
	}

	_, err = db.Exec("INSERT INTO exams (name, date) VALUES (?, ?)", name, date.Format(time.RFC3339))
	if err != nil {
		logger.Error().Err(err).Str("name", name).Msg("Failed to insert exam")
		return err
	}

	logger.Info().Str("name", name).Str("date", dateStr).Msg("Exam added successfully")
	return nil
}

func GetExam(name string) (Exam, error) {
	var exam Exam
	var dateStr string
	err := db.QueryRow("SELECT name, date FROM exams WHERE name = ?", name).Scan(&exam.Name, &dateStr)
	if err != nil {
		logger.Warn().Err(err).Str("name", name).Msg("Exam not found")
		return exam, err
	}

	exam.Date, err = time.Parse(time.RFC3339, dateStr)
	if err != nil {
		logger.Error().Err(err).Str("name", name).Msg("Failed to parse exam date")
		return exam, err
	}

	logger.Info().Str("name", name).Msg("Exam retrieved successfully")
	return exam, nil
}

func DeleteExam(name string) error {
	result, err := db.Exec("DELETE FROM exams WHERE name = ?", name)
	if err != nil {
		logger.Error().Err(err).Str("name", name).Msg("Failed to delete exam")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Str("name", name).Msg("Failed to check rows affected")
		return err
	}

	if rowsAffected == 0 {
		logger.Warn().Str("name", name).Msg("Exam not found for deletion")
		return fmt.Errorf("exam not found")
	}

	logger.Info().Str("name", name).Msg("Exam deleted successfully")
	return nil
}

func UpdateExam(name, dateStr string) error {
	date, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		logger.Error().Err(err).Str("date", dateStr).Msg("Invalid date format")
		return err
	}

	if date.Before(time.Now()) {
		logger.Warn().Str("date", dateStr).Msg("Attempted to update exam to past date")
		return fmt.Errorf("cannot update exam to a past date")
	}

	result, err := db.Exec("UPDATE exams SET date = ? WHERE name = ?", date.Format(time.RFC3339), name)
	if err != nil {
		logger.Error().Err(err).Str("name", name).Msg("Failed to update exam")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Str("name", name).Msg("Failed to check rows affected")
		return err
	}

	if rowsAffected == 0 {
		logger.Warn().Str("name", name).Msg("Exam not found for update")
		return fmt.Errorf("exam not found")
	}

	logger.Info().Str("name", name).Str("date", dateStr).Msg("Exam updated successfully")
	return nil
}

func ListExams() ([]Exam, error) {
	rows, err := db.Query("SELECT name, date FROM exams")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query exams")
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close rows")
		}
	}()

	var exams []Exam
	for rows.Next() {
		var exam Exam
		var dateStr string
		if err := rows.Scan(&exam.Name, &dateStr); err != nil {
			logger.Error().Err(err).Msg("Failed to scan exam row")
			return nil, err
		}
		exam.Date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			logger.Error().Err(err).Str("date", dateStr).Msg("Failed to parse exam date")
			return nil, err
		}
		exams = append(exams, exam)
	}

	if len(exams) == 0 {
		logger.Info().Msg("No exams found")
	} else {
		logger.Info().Int("count", len(exams)).Msg("Exams retrieved successfully")
	}

	return exams, nil
}
