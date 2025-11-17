package src

import (
	"fmt"
	"runtime"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func startHandler(m *telegram.NewMessage) error {
	bot := m.Client.Me()
	response := fmt.Sprintf(`
👋 Hello <b>%s</b>!

Welcome to <b>%s</b> — your assistant to manage Coolify projects.

Use the menu below to get started.`, m.Sender.FirstName, bot.FirstName)

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("📋 List Projects", "list_projects")).
		AddRow(telegram.Button.URL("💫 Fᴀʟʟᴇɴ Pʀᴏᴊᴇᴄᴛꜱ", "https://t.me/FallenProjects")).
		AddRow(telegram.Button.URL("🛠️ Sᴏᴜʀᴄᴇ Cᴏᴅᴇ", "https://github.com/AshokShau/coolify-telegram-bot"))
	_, err := m.Reply(response, &telegram.SendOptions{
		ReplyMarkup: keyboard.Build(),
	})
	return err
}

func pingHandler(m *telegram.NewMessage) error {
	start := time.Now()
	msg, err := m.Reply("⏱️ Pinging...")
	if err != nil {
		return err
	}
	latency := time.Since(start).Milliseconds()
	uptime := time.Since(startTime).Truncate(time.Second)

	response := fmt.Sprintf(
		"<b>📊 System Performance Metrics</b>\n\n"+
			"⏱️ <b>Bot Latency:</b> <code>%d ms</code>\n"+
			"🕒 <b>Uptime:</b> <code>%s</code>\n"+
			"➜ <b>Current Go Routines:</b> <code>%d</code>\n",
		latency, uptime, runtime.NumGoroutine(),
	)

	_, err = msg.Edit(response)
	return err
}
