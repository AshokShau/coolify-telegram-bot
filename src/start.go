package src

import (
	"fmt"
	"runtime"
	"time"

	"github.com/AshokShau/gotdbot"
	"github.com/AshokShau/gotdbot/ext"
)

func startHandler(ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	c := ctx.Client

	response := fmt.Sprintf(`
Welcome to <b>%s</b> — your assistant to manage Coolify projects.
`, c.Me().FirstName)

	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "📋 List Projects",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("list_projects"),
					},
				},
				{
					Text: "💫 Fᴀʟʟᴇɴ Pʀᴏᴊᴇᴄᴛꜱ",
					TypeField: &gotdbot.InlineKeyboardButtonTypeUrl{
						Url: "https://t.me/FallenProjects",
					},
				},
			},
			{
				{
					Text: "🛠 Sᴏᴜʀᴄᴇ Cᴏᴅᴇ",
					TypeField: &gotdbot.InlineKeyboardButtonTypeUrl{
						Url: "https://github.com/AshokShau/coolify-telegram-bot",
					},
				},
			},
		},
	}

	_, err := msg.ReplyText(c, response, &gotdbot.SendTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

func pingHandler(ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	c := ctx.Client

	start := time.Now()
	updateLag := time.Since(time.Unix(int64(msg.Date), 0)).Milliseconds()

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
			"📩 <b>Update Lag:</b> <code>%d ms</code>\n"+
			"⚙️ <b>Go Routines:</b> <code>%d</code>\n",
		latency, uptime, updateLag, runtime.NumGoroutine(),
	)

	_, err = msg.EditText(c, response, &gotdbot.EditTextMessageOpts{ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("failed to edit ping message: %w", err)
	}
	return nil
}
