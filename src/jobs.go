package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"fmt"
	"strings"

	"github.com/AshokShau/gotdbot"
	"github.com/AshokShau/gotdbot/ext"
)

const pageSize = 5

func jobsHandler(ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	c := ctx.Client

	if !config.IsDev(msg.FromID()) {
		_, err := msg.ReplyText(c, "🚫 You are not authorized to use this command.", nil)
		return err
	}

	text, kb, err := buildJobsMessage(1)
	if err != nil {
		_, err = msg.ReplyText(c, "❌ "+err.Error(), nil)
		return err
	}

	_, err = msg.ReplyText(c, text, &gotdbot.SendTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func jobsPaginationHandler(ctx *ext.Context) error {
	c := ctx.Client
	cb := ctx.Update.UpdateNewCallbackQuery
	data := cb.DataString()
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 100)
		return nil
	}

	page := 1
	if parts := strings.Split(data, ":"); len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &page)
	}

	text, kb, err := buildJobsMessage(page)
	if err != nil {
		_ = cb.Answer(c, "Error: "+err.Error(), true, "", 100)
		return nil
	}

	_, err = cb.EditMessageText(c, text, &gotdbot.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func buildJobsMessage(page int) (string, gotdbot.ReplyMarkup, error) {
	tasks, err := database.GetTasks()
	if err != nil {
		return "", nil, fmt.Errorf("error fetching tasks: %v", err)
	}

	if len(tasks) == 0 {
		return "📭 No scheduled jobs found.", nil, nil
	}

	start, end, buttons := Paginate(len(tasks), page, pageSize, "jobs:")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📅 Scheduled Jobs (Page %d):</b>\n\n", page))

	for _, task := range tasks[start:end] {
		sb.WriteString(fmt.Sprintf("🆔 <code>%s</code>\n", task.ID.Hex()))
		sb.WriteString(fmt.Sprintf("🏷️ <b>Name:</b> %s\n", task.Name))
		sb.WriteString(fmt.Sprintf("🔧 <b>Type:</b> %s\n", task.Type))
		sb.WriteString(fmt.Sprintf("⏰ <b>Schedule:</b> %s\n", task.Schedule))
		if task.OneTime {
			sb.WriteString(fmt.Sprintf("⏳ <b>Next Run:</b> %s\n", task.NextRun.Format("2006-01-02 15:04:05")))
		}
		sb.WriteString("➖➖➖➖➖➖➖➖➖➖\n")
	}

	kb := &gotdbot.ReplyMarkupInlineKeyboard{}
	if len(buttons) > 0 {
		row := make([]gotdbot.InlineKeyboardButton, 0, len(buttons))

		for _, btn := range buttons {
			row = append(row, gotdbot.InlineKeyboardButton{
				Text: btn.Text,
				TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
					Data: []byte(btn.Data),
				},
			})
		}

		kb.Rows = append(kb.Rows, row)
	}

	return sb.String(), kb, nil
}
