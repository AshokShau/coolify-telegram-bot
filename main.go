package main

import (
	"coolifymanager/src"
	"coolifymanager/src/config"
	"fmt"
	"log"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

var allowedUpdates = []string{"message", "callback_query"}

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("❌ Failed to initialize config: %v", err)
	}

	bot, err := initBot()
	if err != nil {
		log.Fatalf("❌ Failed to create bot: %v", err)
	}

	updater := ext.NewUpdater(src.Dispatcher, nil)

	if config.WebhookUrl != "" {
		log.Println("🌐 Starting bot in Webhook mode...")
		if err := startWebhookBot(updater, bot, config.WebhookUrl, "super-secret-token"); err != nil {
			log.Fatalf("❌ Webhook init failed: %v", err)
		}
	} else {
		log.Println("📡 Starting bot in Long Polling mode...")
		if err := startLongPollingBot(updater, bot); err != nil {
			log.Fatalf("❌ Polling init failed: %v", err)
		}
	}

	log.Printf("🤖 Bot @%s is now running...\n", bot.User.Username)
	updater.Idle()
}

func initBot() (*gotgbot.Bot, error) {
	bot, err := gotgbot.NewBot(config.Token, nil)
	if err != nil {
		return nil, fmt.Errorf("could not initialize bot: %w", err)
	}
	return bot, nil
}

func startLongPollingBot(updater *ext.Updater, bot *gotgbot.Bot) error {
	return updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout:        9,
			AllowedUpdates: allowedUpdates,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: 10 * time.Second,
			},
		},
	})
}

func startWebhookBot(updater *ext.Updater, bot *gotgbot.Bot, domain, webhookSecret string) error {
	opts := ext.WebhookOpts{
		ListenAddr:  "0.0.0.0:" + config.Port,
		SecretToken: webhookSecret,
	}

	if err := updater.StartServer(opts); err != nil {
		return fmt.Errorf("failed to start webhook server: %w", err)
	}

	if err := updater.AddWebhook(bot, bot.Token, &ext.AddWebhookOpts{
		SecretToken: webhookSecret,
	}); err != nil {
		return fmt.Errorf("failed to add webhook: %w", err)
	}

	if err := updater.SetAllBotWebhooks(domain, &gotgbot.SetWebhookOpts{
		MaxConnections:     100,
		AllowedUpdates:     allowedUpdates,
		DropPendingUpdates: true,
		SecretToken:        webhookSecret,
	}); err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	return nil
}
