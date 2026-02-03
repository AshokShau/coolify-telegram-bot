package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"os"
	"strings"

	"github.com/AshokShau/gotdbot"
	"github.com/AshokShau/gotdbot/ext"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func listProjectsHandler(ctx *ext.Context) error {
	c := ctx.Client
	cb := ctx.Update.UpdateNewCallbackQuery

	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 100)
		return nil
	}
	_ = cb.Answer(c, "Processing...", false, "", 0)
	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_, _ = cb.EditMessageText(c, "Failed to fetch projects:"+err.Error(), nil)
		return nil
	}

	if len(apps) == 0 {
		_, _ = cb.EditMessageText(c, "😶 No applications found.", nil)
		return nil
	}

	page := 1
	dataBytes := cb.CallbackData()
	cbData := string(dataBytes)
	if strings.Contains(cbData, ":") {
		parts := strings.Split(cbData, ":")
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &page)
		}
	}

	start, end, paginationButtons := Paginate(len(apps), page, 7, "list_projects:")

	kb := &gotdbot.ReplyMarkupInlineKeyboard{}
	for _, app := range apps[start:end] {
		text := fmt.Sprintf("📦 %s (%s)", app.Name, app.Status)
		data := "project_menu:" + app.UUID

		kb.Rows = append(kb.Rows, []*gotdbot.InlineKeyboardButton{
			{
				Text: text,
				TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
					Data: []byte(data),
				},
			},
		})
	}

	if len(paginationButtons) > 0 {
		row := make([]*gotdbot.InlineKeyboardButton, 0, len(paginationButtons))

		for _, btn := range paginationButtons {
			row = append(row, &gotdbot.InlineKeyboardButton{
				Text: btn.Text,
				TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
					Data: []byte(btn.Data),
				},
			})
		}

		kb.Rows = append(kb.Rows, row)
	}

	_, err = cb.EditMessageText(c, "<b>📋 Select a project:</b>", &gotdbot.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func projectMenuHandler(ctx *ext.Context) error {
	cb := ctx.Update.UpdateNewCallbackQuery
	c := ctx.Client

	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 0)
		return nil
	}

	_ = cb.Answer(c, "Processing...", false, "", 0)

	uuid := strings.TrimPrefix(string(cb.CallbackData()), "project_menu:")
	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, err = cb.EditMessageText(c, "❌ Failed to load project: "+err.Error(), nil)
		return err
	}

	text := fmt.Sprintf("<b>📦 %s</b>\n🌐 %s\n📄 Status: <code>%s</code>", app.Name, app.FQDN, app.Status)
	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "🔄 Restart",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("restart:" + uuid),
					},
				},
				{
					Text: "🚀 Deploy",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("deploy:" + uuid),
					},
				},
			},
			{
				{
					Text: "📜 Logs",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("logs:" + uuid),
					},
				},
				{
					Text: "ℹ️ Status",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("status:" + uuid),
					},
				},
			},
			{
				{
					Text: "📅 Schedule",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("sch_m:" + uuid),
					},
				},
			},
			{
				{
					Text: "🛑 Stop",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("stop:" + uuid),
					},
				},
				{
					Text: "❌ Delete",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("delete:" + uuid),
					},
				},
			},
			{
				{
					Text: "🔙 Back",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("list_projects:"),
					},
				},
			},
		},
	}

	_, err = cb.EditMessageText(c, text, &gotdbot.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})

	return err
}

func restartHandler(ctx *ext.Context) error {
	cb := ctx.Update.UpdateNewCallbackQuery
	c := ctx.Client
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 0)
		return nil
	}
	_ = cb.Answer(c, "Processing...", false, "", 0)

	uuid := strings.TrimPrefix(string(cb.CallbackData()), "restart:")

	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "🔙 Back",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	res, err := config.Coolify.RestartApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ Restart failed: "+err.Error(), &gotdbot.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	text := fmt.Sprintf("✅ Restart queued!\nDeployment UUID: <code>%s</code>", res.DeploymentUUID)
	_, err = cb.EditMessageText(c, text, &gotdbot.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func deployHandler(ctx *ext.Context) error {
	cb := ctx.Update.UpdateNewCallbackQuery
	c := ctx.Client

	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 0)
		return nil
	}
	_ = cb.Answer(c, "Processing...", false, "", 0)

	uuid := strings.TrimPrefix(string(cb.CallbackData()), "deploy:")

	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "🔙 Back",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	res, err := config.Coolify.StartApplicationDeployment(uuid, false, false)
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ Deploy failed: "+err.Error(), &gotdbot.EditTextMessageOpts{ReplyMarkup: kb})
		return err
	}

	text := fmt.Sprintf("✅ Deployment queued!\nDeployment UUID: <code>%s</code>", res.DeploymentUUID)
	_, err = cb.EditMessageText(c, text, &gotdbot.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func logsHandler(ctx *ext.Context) error {
	cb := ctx.Update.UpdateNewCallbackQuery
	c := ctx.Client

	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 0)
		return nil
	}
	_ = cb.Answer(c, "Processing...", false, "", 0)

	uuid := strings.TrimPrefix(string(cb.CallbackData()), "logs:")

	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "🔙 Back",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	_, _ = cb.EditMessageText(c, "Processing...", nil)
	logsData, err := config.Coolify.GetApplicationLogsByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ Logs error: "+err.Error(), &gotdbot.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	tmpFile, err := os.CreateTemp("", "logs-*.txt")
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ Failed to create temp file: "+err.Error(), nil)
		return err
	}

	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(logsData)); err != nil {
		_, _ = cb.EditMessageText(c, "❌ Failed to write logs: "+err.Error(), nil)
		return err
	}

	tmpFile.Close()

	file := tmpFile.Name()
	_, err = c.EditMessageMedia(cb.ChatId, cb.MessageId, &gotdbot.InputMessageDocument{Document: gotdbot.GetInputFile(file)}, &gotdbot.EditMessageMediaOpts{ReplyMarkup: kb})
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ Failed to send logs file: "+err.Error(), &gotdbot.EditTextMessageOpts{ReplyMarkup: kb})
		return fmt.Errorf("edit message media error: %s", err.Error())
	}

	return nil
}

func statusHandler(ctx *ext.Context) error {
	cb := ctx.Update.UpdateNewCallbackQuery
	c := ctx.Client

	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 0)
		return nil
	}
	_ = cb.Answer(c, "Processing...", false, "", 0)

	uuid := strings.TrimPrefix(string(cb.CallbackData()), "status:")

	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "🔙 Back",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ Status error: "+err.Error(), &gotdbot.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	text := fmt.Sprintf("📦 <b>%s</b>\nCurrent Status: <code>%s</code>", app.Name, app.Status)
	_, err = cb.EditMessageText(c, text, &gotdbot.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func stopHandler(ctx *ext.Context) error {
	cb := ctx.Update.UpdateNewCallbackQuery
	c := ctx.Client

	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 0)
		return nil
	}
	_ = cb.Answer(c, "Processing...", false, "", 0)

	uuid := strings.TrimPrefix(string(cb.CallbackData()), "stop:")

	res, err := config.Coolify.StopApplicationByUUID(uuid)
	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "🔙 Back",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ Stop failed: "+err.Error(), &gotdbot.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	_, err = cb.EditMessageText(c, "🛑 "+res.Message, &gotdbot.EditTextMessageOpts{ReplyMarkup: kb, ParseMode: "HTML"})
	return err
}

func deleteHandler(ctx *ext.Context) error {
	cb := ctx.Update.UpdateNewCallbackQuery
	c := ctx.Client

	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 0)
		return nil
	}
	_ = cb.Answer(c, "Processing...", false, "", 0)

	uuid := strings.TrimPrefix(string(cb.CallbackData()), "delete:")

	err := config.Coolify.DeleteApplicationByUUID(uuid)
	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "🔙 Back",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	if err != nil {
		_, err = cb.EditMessageText(c, "❌ Delete failed: "+err.Error(), &gotdbot.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	_, err = cb.EditMessageText(c, "✅ Application deleted successfully.", &gotdbot.EditTextMessageOpts{ReplyMarkup: kb})
	return err
}

func scheduleMenuHandler(ctx *ext.Context) error {
	cb := ctx.Update.UpdateNewCallbackQuery
	c := ctx.Client

	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 0)
		return nil
	}
	_ = cb.Answer(c, "Processing...", false, "", 0)

	uuid := strings.TrimPrefix(string(cb.CallbackData()), "sch_m:")

	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "🔄 Restart",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("sch_a:" + uuid + ":restart"),
					},
				},
			},
			{
				{
					Text: "🔙 Back",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	_, err := cb.EditMessageText(c, "<b>📅 Select Action Type:</b>", &gotdbot.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}

func scheduleActionHandler(ctx *ext.Context) error {
	cb := ctx.Update.UpdateNewCallbackQuery
	c := ctx.Client

	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 0)
		return nil
	}
	_ = cb.Answer(c, "Processing...", false, "", 0)

	// Format: sch_a:uuid:actionType
	data := strings.TrimPrefix(string(cb.CallbackData()), "sch_a:")
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return nil
	}
	uuid := parts[0]
	actionType := parts[1]

	// Common intervals
	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "Hourly",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_1h", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "Daily",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_1d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "Every 2 Days",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_2d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "Every 3 Days",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_3d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "Weekly",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_7d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "🔙 Back",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("sch_m:" + uuid),
					},
				},
			},
		},
	}

	_, err := cb.EditMessageText(c, "<b>⏰ Select Schedule:</b>", &gotdbot.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}

func scheduleCreateHandler(ctx *ext.Context) error {
	cb := ctx.Update.UpdateNewCallbackQuery
	c := ctx.Client

	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, "🚫 You are not authorized.", true, "", 0)
		return nil
	}
	_ = cb.Answer(c, "Processing...", false, "", 0)

	// Format: sch_c:uuid:actionType:schedule
	data := strings.TrimPrefix(string(cb.CallbackData()), "sch_c:")

	parts := strings.Split(data, ":")
	if len(parts) < 3 {
		return nil
	}
	uuid := parts[0]
	actionType := parts[1]
	schedule := parts[2]

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ Failed to get application: "+err.Error(), nil)
		return nil
	}

	task := database.ScheduledTask{
		ID:          bson.NewObjectID(),
		Name:        app.Name,
		ProjectUUID: uuid,
		Type:        actionType,
		Schedule:    schedule,
	}

	if err := database.AddTask(task); err != nil {
		_, _ = cb.EditMessageText(c, "❌ Failed to save task: "+err.Error(), nil)
		return nil
	}

	if err := scheduler.ScheduleTask(task); err != nil {
		_ = database.DeleteTask(task.ID.Hex())
		_, _ = cb.EditMessageText(c, "❌ Failed to schedule task: "+err.Error(), nil)
		return nil
	}

	kb := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]*gotdbot.InlineKeyboardButton{
			{
				{
					Text: "🔙 Back",
					TypeField: &gotdbot.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	_, err = cb.EditMessageText(c, fmt.Sprintf("✅ Task scheduled successfully!\n\nID: <code>%s</code>\nType: %s\nSchedule: %s", task.ID.Hex(), actionType, schedule), &gotdbot.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}
