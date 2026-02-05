package config

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"coolifymanager/src/coolity"
	"coolifymanager/src/database"
)

var (
	Coolify          *coolify.Client
	Token            = os.Getenv("TOKEN")
	ApiId            = os.Getenv("API_ID")
	ApiHash          = os.Getenv("API_HASH")
	apiUrl           = os.Getenv("API_URL")
	apiToken         = os.Getenv("API_TOKEN")
	devList          = os.Getenv("DEV_IDS")
	dbURL            = os.Getenv("DB_URL")
	TdlibLibraryPath = os.Getenv("TDLIB_LIBRARY_PATH")
	devIDs           []int64
)

// loadEnvFile loads environment variables from a file
func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentKey string
	var currentValue strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if currentKey != "" && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
			currentValue.WriteString("\n" + strings.TrimSpace(line))
			continue
		}

		if currentKey != "" {
			value := strings.TrimSpace(currentValue.String())
			os.Setenv(currentKey, unquoteValue(value))
			currentKey = ""
			currentValue.Reset()
		}

		idx := strings.Index(line, "=")
		if idx == -1 {
			log.Printf("Skipping invalid line in .env: %s", line)
			continue
		}

		key := strings.TrimSpace(line[:idx])
		valuePart := strings.TrimSpace(line[idx+1:])
		if commentIdx := strings.Index(valuePart, " #"); commentIdx != -1 {
			valuePart = strings.TrimSpace(valuePart[:commentIdx])
		}

		if strings.HasSuffix(valuePart, "\\") {
			currentKey = key
			currentValue.WriteString(strings.TrimSuffix(valuePart, "\\"))
			continue
		}

		os.Setenv(key, unquoteValue(valuePart))
	}

	if currentKey != "" {
		value := strings.TrimSpace(currentValue.String())
		os.Setenv(currentKey, unquoteValue(value))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	return nil
}

// unquoteValue removes surrounding quotes from values
func unquoteValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}

	return value
}

// Load multiple env files in order, later files override earlier ones
func loadEnvFiles(paths ...string) error {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if err := loadEnvFile(path); err != nil {
				return err
			}
		}
	}
	return nil
}

func InitConfig() error {
	if err := loadEnvFiles(".env"); err != nil {
		log.Printf("Warning: %v", err)
	}

	// Re-read environment variables after loading .env file
	reloadEnvVars()

	if err := validateRequiredEnv(); err != nil {
		return err
	}

	// Initialize HTTP client
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// Initialize Coolify client
	Coolify = coolify.NewClient(
		apiUrl,
		apiToken,
		httpClient,
		30*time.Minute,
	)

	// Parse DEV_IDS
	if err := parseDevIDs(); err != nil {
		return fmt.Errorf("error parsing DEV_IDS: %w", err)
	}

	// Initialize Database
	if err := database.Connect(dbURL); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Configuration initialized successfully")
	return nil
}

// reloadEnvVars reloads environment variables from os.Getenv
func reloadEnvVars() {
	Token = os.Getenv("TOKEN")
	ApiId = os.Getenv("API_ID")
	ApiHash = os.Getenv("API_HASH")
	apiUrl = os.Getenv("API_URL")
	apiToken = os.Getenv("API_TOKEN")
	devList = os.Getenv("DEV_IDS")
	dbURL = os.Getenv("DB_URL")
	TdlibLibraryPath = os.Getenv("TDLIB_LIBRARY_PATH")
}

// validateRequiredEnv checks all required environment variables
func validateRequiredEnv() error {
	required := map[string]string{
		"API_URL":   apiUrl,
		"API_TOKEN": apiToken,
		"TOKEN":     Token,
	}

	var missing []string
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// parseDevIDs parses the DEV_IDS environment variable
func parseDevIDs() error {
	devIDs = []int64{} // Reset slice

	if devList == "" {
		return nil // No dev IDs is valid
	}

	for _, idStr := range strings.Split(devList, ",") {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid DEV_ID '%s': %w", idStr, err)
		}
		devIDs = append(devIDs, id)
	}

	return nil
}

// IsDev checks if a given Telegram user ID is in the dev list
func IsDev(userID int64) bool {
	for _, id := range devIDs {
		if id == userID {
			return true
		}
	}
	return false
}

// GetEnv returns environment variable with default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEnvInt returns environment variable as int with default value
func GetEnvInt(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}

// GetEnvBool returns environment variable as bool with default value
func GetEnvBool(key string, defaultValue bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.ParseBool(value)
}
