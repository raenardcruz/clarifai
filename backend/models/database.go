package models

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func GetDB() *gorm.DB {
	return DB
}

func EnsureDatabaseExists() {
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "admin"
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "admin"
	}
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "clarifi_db"
	}

	// Connect to default 'postgres' database to ensure our DB exists
	dsnDefault := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable", host, user, password, port)
	dbDefault, err := gorm.Open(postgres.Open(dsnDefault), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Failed to connect to default postgres DB: %v. Proceeding to connect to target DB directly.", err)
		return
	}

	// Check if DB exists
	var exists int
	err = dbDefault.Raw("SELECT 1 FROM pg_database WHERE datname = ?", dbName).Scan(&exists).Error
	if err != nil {
		log.Printf("Warning: failed to query pg_database: %v", err)
		return
	}

	if exists == 0 {
		log.Printf("Creating database %s...", dbName)
		// CREATE DATABASE cannot run inside a transaction block, GORMExec executes it raw
		err = dbDefault.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error
		if err != nil {
			log.Printf("Warning: failed to create database %s: %v", dbName, err)
		}
	}
}

func InitDB() {
	EnsureDatabaseExists()

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "admin"
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "admin"
	}
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "clarifi_db"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", host, user, password, dbName, port)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established.")

	// Auto Migration
	log.Println("Running AutoMigration...")

	// Migrate legacy GORM column 'a_iprovider' to 'ai_provider' if it was created
	_ = DB.Exec("DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='settings' AND column_name='a_iprovider') AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='settings' AND column_name='ai_provider') THEN ALTER TABLE settings RENAME COLUMN a_iprovider TO ai_provider; END IF; END $$;").Error

	err = DB.AutoMigrate(&User{}, &Recording{}, &TranscriptSegment{}, &Settings{}, &Tag{}, &MeetingInvite{}, &RecordingAccess{})
	if err != nil {
		log.Fatalf("Failed to run AutoMigration: %v", err)
	}

	// In case both existed, migrate data and drop legacy column
	_ = DB.Exec("DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='settings' AND column_name='a_iprovider') THEN UPDATE settings SET ai_provider = a_iprovider WHERE (ai_provider IS NULL OR ai_provider = '') AND a_iprovider IS NOT NULL; ALTER TABLE settings DROP COLUMN a_iprovider; END IF; END $$;").Error

	// Seeding
	SeedDB()
}

func SeedDB() {
	// Check/create admin user
	var count int64
	DB.Model(&User{}).Where("email = ?", "admin").Count(&count)
	if count == 0 {
		log.Println("Creating default admin user...")
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing admin password: %v", err)
			return
		}
		admin := User{
			Email:          "admin",
			HashedPassword: string(hashedBytes),
			Role:           "admin",
			IsApproved:     true,
		}
		if err := DB.Create(&admin).Error; err != nil {
			log.Printf("Failed to create admin: %v", err)
		} else {
			log.Println("Admin user created (admin/admin).")
		}
	}

	// Check/create default settings
	DB.Model(&Settings{}).Count(&count)
	if count == 0 {
		log.Println("Creating default settings...")
		defaultSettings := Settings{
			MedicalSummaryPrompt: DefaultMedicalSummaryPrompt,
			AIProvider:           "ollama",
			OllamaURL:            "http://localhost:11434",
			OllamaModel:          "gemma4:12b-mlx",
			GeminiModel:          "gemini-3.8-flash",
		}
		if err := DB.Create(&defaultSettings).Error; err != nil {

			log.Printf("Failed to create default settings: %v", err)
		} else {
			log.Println("Default settings created.")
		}
	} else {
		// Update empty/null fields in existing settings
		DB.Model(&Settings{}).Where("ollama_url = ? OR ollama_url IS NULL", "").Update("ollama_url", "http://localhost:11434")
		DB.Model(&Settings{}).Where("ollama_model = ? OR ollama_model IS NULL", "").Update("ollama_model", "gemma4:12b-mlx")
		DB.Model(&Settings{}).Where("ai_provider = ? OR ai_provider IS NULL", "").Update("ai_provider", "ollama")
		DB.Model(&Settings{}).Where("gemini_model = ? OR gemini_model IS NULL", "").Update("gemini_model", "gemini-3.8-flash")
		DB.Model(&Settings{}).Where("medical_summary_prompt = ? OR medical_summary_prompt IS NULL", "").Update("medical_summary_prompt", DefaultMedicalSummaryPrompt)
	}
}

