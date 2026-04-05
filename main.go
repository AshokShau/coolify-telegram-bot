package main

//go:generate go run github.com/AshokShau/gotdbot/scripts/tools@latest

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
		log.Fatalf("❌ Failed to initialize config: %v", err)
	}

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Printf("⚠️ Failed to load Asia/Kolkata time zone: %v. Using UTC.", err)
	} else {
		time.Local = loc
		log.Println("✅ Time zone set to Asia/Kolkata")
	}

	apiID, err := strconv.Atoi(config.ApiId)
	if err != nil {
		log.Fatalf("❌ Invalid API_ID: %v", err)
	}

	tdlibLibraryPath := config.TdlibLibraryPath
	if tdlibLibraryPath == "" {
		tdlibLibraryPath = "./libtdjson.so.1.8.63"
	}

	bot, err := gotdbot.NewClient(int32(apiID), config.ApiHash, config.Token, &gotdbot.ClientOpts{LibraryPath: tdlibLibraryPath})
	if err != nil {
		log.Fatalf("❌ Failed to create bot client: %v", err)
	}
	err = src.InitFunc(bot.Dispatcher)
	if err != nil {
		panic(err.Error())
	}

	if err = bot.Start(); err != nil {
		panic(err.Error())
	}

	me := bot.Me
	username := ""
	if me.Usernames != nil && len(me.Usernames.ActiveUsernames) > 0 {
		username = me.Usernames.ActiveUsernames[0]
	}

	bot.Logger.Info("✅ Bot started as @" + username)
	bot.Idle()
}
