package src

import (
	"fmt"
	"runtime"
	"time"

	td "github.com/AshokShau/gotdbot"
)

func startHandler(c *td.Client, msg *td.Message) error {
	response := fmt.Sprintf("Welcome to <b>%s</b> — your assistant to manage Coolify projects.", c.Me.FirstName)

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "List Projects",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("list_projects"),
					},
				},
				{
					Text: "Fallen Projects",
					Type: &td.InlineKeyboardButtonTypeUrl{
						Url: "https://t.me/FallenProjects",
					},
				},
			},
			{
				{
					Text: "Source Code",
					Type: &td.InlineKeyboardButtonTypeUrl{
						Url: "https://github.com/AshokShau/coolify-telegram-bot",
					},
				},
			},
		},
	}

	_, err := replyMsg(c, msg, response, &td.SendTextMessageOpts{ReplyMarkup: kb})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

func pingHandler(c *td.Client, msg *td.Message) error {
	start := time.Now()
	msg, err := msg.ReplyText(c, "⏱️ Pinging...", nil)
	if err != nil {
		return fmt.Errorf("failed to send ping message: %w", err)
	}

	latency := time.Since(start).Milliseconds()
	uptime := time.Since(startTime).Truncate(time.Second)

	response := fmt.Sprintf(
		"<b>📊 System Performance Metrics</b>\n\n"+
			"⏱️ <b>Bot Latency:</b> <code>%d ms</code>\n"+
			"🕒 <b>Uptime:</b> <code>%s</code>\n"+
			"⚙️ <b>Go Routines:</b> <code>%d</code>\n",
		latency, uptime, runtime.NumGoroutine(),
	)

	_, err = msg.EditText(c, response, &td.EditTextMessageOpts{ParseMode: "HTML"})
	return err
}
