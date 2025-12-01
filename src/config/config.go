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
	Coolify  *coolify.Client
	Token    = os.Getenv("TOKEN")
	ApiId    = os.Getenv("API_ID")
	ApiHash  = os.Getenv("API_HASH")
	apiUrl   = os.Getenv("API_URL")
	apiToken = os.Getenv("API_TOKEN")
	devList  = os.Getenv("DEV_IDS")
	devIDs   []int64
)

func InitConfig() error {
	if apiUrl == "" || apiToken == "" || Token == "" {
		return errors.New("API_URL , API_TOKEN and TOKEN must be set")
	}

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}
	// Initialize Coolify client with 30-minute cache TTL
	Coolify = coolify.NewClient(
		apiUrl,
		apiToken,
		httpClient,
		30*time.Minute,
	)

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
