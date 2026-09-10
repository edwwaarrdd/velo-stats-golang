package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv string

	HTTPPort           string
	CORSAllowedOrigins []string

	DatabasePath string

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	QueuePrefix   string

	StationInformationURL string
	OSRMBaseURL           string
	OpenMeteoArchiveURL   string
	RidesJSONPath         string
}

func Load() Config {
	loadDotEnv(".env")

	return Config{
		AppEnv: env("APP_ENV", "local"),

		HTTPPort:           env("HTTP_PORT", "8000"),
		CORSAllowedOrigins: splitAndTrim(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),

		DatabasePath: env("DB_DATABASE", "database/database.sqlite"),

		RedisAddr:     env("REDIS_HOST", "127.0.0.1") + ":" + env("REDIS_PORT", "6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RedisDB:       envInt("REDIS_DB", 0),
		QueuePrefix:   env("QUEUE_PREFIX", "velo-stats:queues:"),

		StationInformationURL: env("VELO_ANTWERP_STATION_INFORMATION_URL", "https://gbfs.smartbike.com/antwerp/1.0/en/station_information.json"),
		OSRMBaseURL:           env("OSRM_BASE_URL", "https://routing.openstreetmap.de"),
		OpenMeteoArchiveURL:   env("OPEN_METEO_ARCHIVE_URL", "https://archive-api.open-meteo.com/v1/archive"),
		RidesJSONPath:         env("RIDES_JSON_PATH", "data/rides.json"),
	}
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, ""))
	if err != nil {
		return fallback
	}

	return value
}

func splitAndTrim(value string) []string {
	var parts []string

	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}

	return parts
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}
