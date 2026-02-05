package main

//go:generate go run github.com/AshokShau/gotdbot/scripts/tools@latest

import (
	"coolifymanager/src"
	"coolifymanager/src/config"
	"log"
	"strconv"

	"github.com/AshokShau/gotdbot"
	"github.com/AshokShau/gotdbot/ext"
)

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatalf("❌ Failed to initialize config: %v", err)
	}

	apiID, err := strconv.Atoi(config.ApiId)
	if err != nil {
		log.Fatalf("❌ Invalid API_ID: %v", err)
	}

	tdlibLibraryPath := config.TdlibLibraryPath
	if tdlibLibraryPath == "" {
		tdlibLibraryPath = "./libtdjson.so.1.8.60"
	}

	bot := gotdbot.NewClient(int32(apiID), config.ApiHash, config.Token, &gotdbot.ClientConfig{LibraryPath: "./libtdjson.so.1.8.60"})

	// gotdbot.SetTdlibLogStreamFile("tdlib.log", 10*1024*1024, false)
	// disable tdlib logging
	gotdbot.SetTdlibLogStreamEmpty()

	dispatcher := ext.NewDispatcher(bot)

	err = src.InitFunc(dispatcher)
	if err != nil {
		panic(err.Error())
	}

	dispatcher.Start()
	if err = bot.Start(); err != nil {
		panic(err.Error())
	}

	me := bot.Me()
	username := ""
	if me.Usernames != nil && len(me.Usernames.ActiveUsernames) > 0 {
		username = me.Usernames.ActiveUsernames[0]
	}

	bot.Logger.Info("✅ Bot started as @" + username)
	bot.Idle()
}
