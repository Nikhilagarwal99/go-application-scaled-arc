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

	// Slave — handles all reads
	// Add more slaves by adding DBSlave2Host etc.
	DBSlaveHost     string
	DBSlavePort     string
	DBSlaveUser     string
	DBSlavePassword string
	DBSlaveName     string
	DBSlaveSSLMode  string

	JWTSecret      string
	JWTExpiryHours int

	MailjetApiKey      string //mailjetAPI_KEY
	MailjetSecret      string //mailjetAPI_SECRET'
	MailjetSenderEmail string
	MailjetSenderName  string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	AWSRegion    string
	AWSAccessKey string
	AWSSecretKey string
	AWSBucket    string
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

		// Slave — falls back to master values if not set.
		// This means in development with one DB, everything
		// still works without any slave configured.
		DBSlaveHost:     getEnv("DB_SLAVE_HOST", getEnv("DB_HOST", "localhost")),
		DBSlavePort:     getEnv("DB_SLAVE_PORT", getEnv("DB_PORT", "5432")),
		DBSlaveUser:     getEnv("DB_SLAVE_USER", getEnv("DB_USER", "postgres")),
		DBSlavePassword: getEnv("DB_SLAVE_PASSWORD", getEnv("DB_PASSWORD", "")),
		DBSlaveName:     getEnv("DB_SLAVE_NAME", getEnv("DB_NAME", "goapp_db")),
		DBSlaveSSLMode:  getEnv("DB_SLAVE_SSLMODE", getEnv("DB_SSLMODE", "disable")),

		JWTSecret:      getEnv("JWT_SECRET", "change-me"),
		JWTExpiryHours: jwtExpiry,

		MailjetApiKey:      getEnv("MAILJET_API_KEY", ""),
		MailjetSecret:      getEnv("MAILJET_API_SECRET", ""),
		MailjetSenderEmail: getEnv("MAILJET_SENDER_EMAIL", ""),
		MailjetSenderName:  getEnv("MAILJET_SENDER_NAME", ""),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		// AWS
		AWSRegion:    getEnv("AWS_REGION", "ap-south-1"),
		AWSAccessKey: getEnv("AWS_ACCESS_KEY", ""),
		AWSSecretKey: getEnv("AWS_SECRET_KEY", ""),
		AWSBucket:    getEnv("AWS_BUCKET", ""),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
