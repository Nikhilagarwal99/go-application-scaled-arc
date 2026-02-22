package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv     string
	ServerPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret      string
	JWTExpiryHours int

	MailjetApiKey      string //mailjetAPI_KEY
	MailjetSecret      string //mailjetAPI_SECRET'
	MailjetSenderEmail string
	MailjetSenderName  string

	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

// Load reads environment variables and returns a Config struct.
// It attempts to load a .env file but does not fail if one is absent
// (useful in container environments where vars are injected directly).
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — reading environment variables directly")
	}

	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		jwtExpiry = 24
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		AppEnv:     getEnv("APP_ENV", "development"),
		ServerPort: getEnv("SERVER_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "goapp_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:      getEnv("JWT_SECRET", "change-me"),
		JWTExpiryHours: jwtExpiry,

		MailjetApiKey:      getEnv("MAILJET_API_KEY", ""),
		MailjetSecret:      getEnv("MAILJET_API_SECRET", ""),
		MailjetSenderEmail: getEnv("MAILJET_SENDER_EMAIL", ""),
		MailjetSenderName:  getEnv("MAILJET_SENDER_NAME", ""),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
