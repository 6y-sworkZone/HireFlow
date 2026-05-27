package config

import (
	"os"
)

type Config struct {
	Port        string
	DBPath      string
	JWTSecret   string
	UploadDir   string
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	SMTPFrom    string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8765"),
		DBPath:    getEnv("DB_PATH", "./data/hireflow.db"),
		JWTSecret: getEnv("JWT_SECRET", "hireflow-jwt-secret-key-change-in-production"),
		UploadDir: getEnv("UPLOAD_DIR", "./uploads"),
		SMTPHost:  getEnv("SMTP_HOST", ""),
		SMTPPort:  getEnv("SMTP_PORT", "587"),
		SMTPUser:  getEnv("SMTP_USER", ""),
		SMTPPass:  getEnv("SMTP_PASS", ""),
		SMTPFrom:  getEnv("SMTP_FROM", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
