package config

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"coolifymanager/src/coolity"
	_ "github.com/joho/godotenv/autoload"
)

var (
	Coolify    *coolify.Client
	ApiUrl     = os.Getenv("API_URL")
	ApiToken   = os.Getenv("API_TOKEN")
	Token      = os.Getenv("TOKEN")
	Port       = os.Getenv("PORT")
	WebhookUrl = os.Getenv("WEBHOOK_URL")
	devList    = os.Getenv("DEV_IDS") // comma-separated
	devIDs     []int64                // parsed slice
)

func Init() error {
	if ApiUrl == "" || ApiToken == "" {
		return errors.New("API_URL and API_TOKEN must be set")
	}

	Coolify = &coolify.Client{
		BaseURL: ApiUrl,
		Token:   ApiToken,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Parse DEV_IDS
	for _, idStr := range strings.Split(devList, ",") {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			devIDs = append(devIDs, id)
		} else {
			log.Printf("Dev ID is not an integer: %s", idStr)
		}
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
