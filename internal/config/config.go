package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	Env                string
	DatabaseURL        string
	SupabaseURL        string
	SupabaseAnonKey    string
	SupabaseServiceKey string
	CORSAllowedOrigins string
	JWTSecret          string
}

// LoadConfig reads configuration from environment variables, optionally loading a .env file.
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] No .env file found or failed to load, reading system environment variables")
	}

	port := getEnv("PORT", "8080")
	env := getEnv("APP_ENV", "production")
	dbURL := getEnv("DATABASE_URL", "")
	supabaseURL := getEnv("SUPABASE_URL", "")
	supabaseAnonKey := getEnv("SUPABASE_ANON_KEY", "")
	supabaseServiceKey := getEnv("SUPABASE_SERVICE_ROLE_KEY", "")
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "*")
	jwtSecret := getEnv("JWT_SECRET", "blindtag-secure-token-secret-key-2026-runner")

	return &Config{
		Port:               port,
		Env:                env,
		DatabaseURL:        dbURL,
		SupabaseURL:        supabaseURL,
		SupabaseAnonKey:    supabaseAnonKey,
		SupabaseServiceKey: supabaseServiceKey,
		CORSAllowedOrigins: corsOrigins,
		JWTSecret:          jwtSecret,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
