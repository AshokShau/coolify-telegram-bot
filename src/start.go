package src

import (
	"fmt"
	"html"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func startHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	startText := fmt.Sprintf(
		`👋 Hello <b>%s</b>!

Welcome to <b>CoolifyBot</b> — your assistant to manage Coolify projects.

Use the menu below to get started.`,
		html.EscapeString(ctx.EffectiveUser.FirstName),
	)

	startMarkup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "📋 List Projects", CallbackData: "list_projects"},
			},
			{
				{Text: "🆘 Support Chat", Url: "https://t.me/GuardxSupport"},
				{Text: "📣 Updates", Url: "https://t.me/FallenProjects"},
			},
		},
	}

	opts := &gotgbot.SendMessageOpts{
		ParseMode:          "HTML",
		ReplyMarkup:        startMarkup,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{IsDisabled: true},
	}

	if _, err := msg.Reply(b, startText, opts); err != nil {
		return err
	}

	return ext.EndGroups
}

func pingCommandHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	start := time.Now()
	msg, err := ctx.EffectiveMessage.Reply(b, "🏓 Pinging...", nil)
	if err != nil {
		return fmt.Errorf("ping: failed to send initial message: %w", err)
	}

	latency := time.Since(start).Milliseconds()
	uptime := time.Since(startTime).Truncate(time.Second)

	response := fmt.Sprintf(
		"<b>📊 System Performance Metrics</b>\n\n"+
			"⏱️ <b>Bot Latency:</b> <code>%d ms</code>\n"+
			"🕒 <b>Uptime:</b> <code>%s</code>\n",
		latency, uptime,
	)

	_, _, err = msg.EditText(b, response, &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
	})
	if err != nil {
		return fmt.Errorf("ping: failed to edit message: %w", err)
	}
	return nil
}
