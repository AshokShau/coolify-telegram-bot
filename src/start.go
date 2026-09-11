package src

import (
	"fmt"
	"runtime"
	"time"

	td "github.com/AshokShau/gotdbot"
)

func getStartContent(c *td.Client) (string, td.ReplyMarkup) {
	text := fmt.Sprintf("Welcome to <b>%s</b> — your assistant to manage Coolify projects.", c.Me.FirstName)

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				makeCallbackButton("Projects", "list_projects:projects"),
				makeCallbackButton("Applications", "list_projects:"),
			},
			{
				makeCallbackButton("Scheduled Jobs", "jobs:1"),
				makeCallbackButton("Deployments", "deployments:"),
			},
			{
				makeCallbackButton("Servers", "servers:"),
				makeCallbackButton("Databases", "databases:"),
			},
			{
				makeCallbackButton("Services", "services:"),
				makeCallbackButton("Team", "team_info:"),
				makeCallbackButton("Tags", "tags:"),
			},
			{
				{
					Text: "Fallen Projects",
					Type: &td.InlineKeyboardButtonTypeUrl{
						Url: "https://t.me/FallenProjects",
					},
				},
				{
					Text: "Source Code",
					Type: &td.InlineKeyboardButtonTypeUrl{
						Url: "https://github.com/AshokShau/coolify-telegram-bot",
					},
				},
			},
		},
	}
	return text, kb
}

func startHandler(c *td.Client, msg *td.Message) error {
	text, kb := getStartContent(c)
	_, err := msg.ReplyText(c, text, sendOpts(msg, kb))
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

func startMenuCallbackHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")
	text, kb := getStartContent(c)
	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func pingHandler(c *td.Client, msg *td.Message) error {
	start := time.Now()
	msg, err := msg.ReplyText(c, "Pinging...", nil)
	if err != nil {
		return fmt.Errorf("failed to send ping message: %w", err)
	}

	latency := time.Since(start).Milliseconds()
	uptime := time.Since(startTime).Truncate(time.Second)

	response := fmt.Sprintf(
		"<b>System Performance Metrics</b>\n\n"+
			"<b>Bot Latency:</b> <code>%d ms</code>\n"+
			"<b>Uptime:</b> <code>%s</code>\n"+
			"<b>Go Routines:</b> <code>%d</code>\n",
		latency, uptime, runtime.NumGoroutine(),
	)

	_, err = msg.EditText(c, response, nil)
	return err
}
