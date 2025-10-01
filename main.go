package main

import (
	"coolifymanager/src"
	"coolifymanager/src/config"
	"log"
	_ "net/http"
	_ "net/http/pprof"
	"strconv"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

// handleFlood delays on flood wait errors
func handleFlood(err error) bool {
	if wait := tg.GetFloodWait(err); wait > 0 {
		log.Printf("⚠️ Flood wait detected: sleeping for %ds", wait)
		time.Sleep(time.Duration(wait) * time.Second)
		return true
	}
	return false
}

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatalf("❌ Failed to initialize config: %v", err)
	}

	apiId, err := strconv.Atoi(config.ApiId)
	if err != nil {
		log.Fatalf("❌ Invalid API_ID: %v", err)
	}

	cfg := tg.NewClientConfigBuilder(int32(apiId), config.ApiHash).
		WithSession("coolify.dat").
		WithLogger(tg.NewLogger(tg.LogInfo).NoColor()).
		WithFloodHandler(handleFlood).
		Build()

	client, err := tg.NewClient(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}
	_, err = client.Conn()
	if err != nil {
		log.Fatalf("❌ Failed to connect to Telegram: %v", err)
	}

	err = client.LoginBot(config.Token)
	if err != nil {
		log.Fatalf("❌ Failed to login bot: %v", err)
	}

	src.InitFunc(client)
	client.Logger.Info("Bot is running as @" + client.Me().Username)
	client.Idle()
}
