package main

//go:generate go run github.com/AshokShau/gotdbot/scripts/tools

import (
	"coolifymanager/src"
	"coolifymanager/src/config"
	"log"
	"strconv"
	"time"
	_ "time/tzdata"

	"github.com/AshokShau/gotdbot"
)

func main() {
	if err := config.InitConfig(); err != nil {
		panic("failed to init config" + err.Error())
	}

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Printf("Failed to load Asia/Kolkata time zone: %v. Using UTC.", err)
	} else {
		time.Local = loc
	}

	apiID, err := strconv.Atoi(config.ApiId)
	if err != nil {
		log.Fatalf("Invalid API_ID: %v", err)
	}

	tdlibLibraryPath := "./libtdjson.so.1.8.66"
	bot, err := gotdbot.NewClient(int32(apiID), config.ApiHash, config.Token, &gotdbot.ClientOpts{
		LibraryPath: tdlibLibraryPath,
		ParseMode:   gotdbot.ParseModeHTML,
		AutoRetry:   &gotdbot.AutoRetry{MaxFloodWait: 5 * time.Minute, ChatNotFound: true},
	})

	if err != nil {
		log.Fatalf("Failed to create bot client: %v", err)
	}
	err = src.InitFunc(bot)
	if err != nil {
		panic("failed to initialize bot: " + err.Error())
	}

	if err = bot.Start(); err != nil {
		panic("failed to start bot: " + err.Error())
	}

	bot.Idle()
}
