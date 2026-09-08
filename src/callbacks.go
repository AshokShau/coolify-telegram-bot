package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type DeleteConfirmStatus int

const (
	DeleteConfirmFirstClick DeleteConfirmStatus = iota
	DeleteConfirmTooSoon
	DeleteConfirmExpired
	DeleteConfirmProceed
)

var deletePendingMap sync.Map // map[string]time.Time

func checkDeleteConfirmation(userID int64, uuid string, now time.Time) DeleteConfirmStatus {
	key := fmt.Sprintf("%d:%s", userID, uuid)
	val, loaded := deletePendingMap.Load(key)

	if !loaded {
		deletePendingMap.Store(key, now)
		return DeleteConfirmFirstClick
	}

	firstClick := val.(time.Time)
	elapsed := now.Sub(firstClick)

	if elapsed < 2*time.Second {
		return DeleteConfirmTooSoon
	}

	if elapsed > 30*time.Second {
		deletePendingMap.Store(key, now)
		return DeleteConfirmExpired
	}

	deletePendingMap.Delete(key)
	return DeleteConfirmProceed
}

func listProjectsHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}

	_ = cb.Answer(c, 0, false, "Processing...", "")
	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_, _ = cb.EditMessageText(c, "Failed to fetch projects:"+err.Error(), nil)
		return nil
	}

	if len(apps) == 0 {
		_, _ = cb.EditMessageText(c, "No applications found.", nil)
		return nil
	}

	page := 1
	cbData := cb.DataString()
	if strings.Contains(cbData, ":") {
		parts := strings.Split(cbData, ":")
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &page)
		}
	}

	start, end, paginationButtons := Paginate(len(apps), page, 7, "list_projects:")

	kb := &td.ReplyMarkupInlineKeyboard{}
	for _, app := range apps[start:end] {
		text := fmt.Sprintf("📦 %s (%s)", app.Name, app.Status)
		data := "project_menu:" + app.UUID

		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			{
				Text: text,
				Type: &td.InlineKeyboardButtonTypeCallback{
					Data: []byte(data),
				},
			},
		})
	}

	if len(paginationButtons) > 0 {
		row := make([]td.InlineKeyboardButton, 0, len(paginationButtons))

		for _, btn := range paginationButtons {
			row = append(row, td.InlineKeyboardButton{
				Text: btn.Text,
				Type: &td.InlineKeyboardButtonTypeCallback{
					Data: []byte(btn.Data),
				},
			})
		}

		kb.Rows = append(kb.Rows, row)
	}

	_, err = cb.EditMessageText(c, "<b>📋 Select a project:</b>", &td.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func projectMenuHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}

	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "project_menu:")
	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, err = cb.EditMessageText(c, "❌ Failed to load project: "+err.Error(), nil)
		return err
	}

	text := fmt.Sprintf("<b>📦 %s</b>\n🌐 %s\n📄 Status: <code>%s</code>", app.Name, app.FQDN, app.Status)
	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔄 Restart",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("restart:" + uuid),
					},
				},
				{
					Text: "🚀 Deploy",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("deploy:" + uuid),
					},
				},
			},
			{
				{
					Text: "📜 Logs",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("logs:" + uuid),
					},
				},
				{
					Text: "ℹ️ Status",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("status:" + uuid),
					},
				},
			},
			{
				{
					Text: "📅 Schedule",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("sch_m:" + uuid),
					},
				},
			},
			{
				{
					Text: "🛑 Stop",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("stop:" + uuid),
					},
				},
				{
					Text: "❌ Delete",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("delete:" + uuid),
					},
				},
			},
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("list_projects:"),
					},
				},
			},
		},
	}

	_, err = cb.EditMessageText(c, text, &td.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})

	return err
}

func restartHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "restart:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	res, err := config.Coolify.RestartApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "Restart failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	text := "Restart queued!"
	if res.DeploymentUUID != "" {
		text += fmt.Sprintf("\nDeployment UUID: <code>%s</code>", res.DeploymentUUID)
	} else if res.Message != "" {
		text += fmt.Sprintf("\nMessage: %s", res.Message)
	}
	_, err = cb.EditMessageText(c, text, &td.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func deployHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}

	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "deploy:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	res, err := config.Coolify.StartApplicationDeployment(uuid, false, false)
	if err != nil {
		_, _ = cb.EditMessageText(c, "Deploy failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return err
	}

	text := fmt.Sprintf("Deployment queued!\nDeployment UUID: <code>%s</code>", res.DeploymentUUID)
	_, err = cb.EditMessageText(c, text, &td.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func logsHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "logs:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	_, _ = cb.EditMessageText(c, "Processing...", nil)
	logsData, err := config.Coolify.GetApplicationLogsByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "Logs error: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	tmpFile, err := os.CreateTemp("", "logs-*.txt")
	if err != nil {
		_, _ = cb.EditMessageText(c, "Failed to create temp file: "+err.Error(), nil)
		return err
	}

	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(logsData)); err != nil {
		_, _ = cb.EditMessageText(c, "Failed to write logs: "+err.Error(), nil)
		return err
	}

	tmpFile.Close()

	file := tmpFile.Name()
	_, err = c.EditMessageMedia(cb.ChatId, &td.InputMessageDocument{Document: &td.InputDocument{Document: td.InputFileLocal{Path: file}}}, cb.MessageId, &td.EditMessageMediaOpts{ReplyMarkup: kb})
	if err != nil {
		_, _ = cb.EditMessageText(c, "Failed to send logs file: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return fmt.Errorf("edit message media error: %s", err.Error())
	}

	return nil
}

func statusHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}
	_ = cb.Answer(c, 0, true, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "status:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "Status error: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	text := fmt.Sprintf("<b>%s</b>\nCurrent Status: <code>%s</code>", app.Name, app.Status)
	_, err = cb.EditMessageText(c, text, &td.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func stopHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "stop:")

	res, err := config.Coolify.StopApplicationByUUID(uuid)
	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	if err != nil {
		_, _ = cb.EditMessageText(c, "Stop failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	_, err = cb.EditMessageText(c, res.Message, &td.EditTextMessageOpts{ReplyMarkup: kb, ParseMode: "HTML"})
	return err
}

func deleteHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "delete:")

	status := checkDeleteConfirmation(cb.SenderUserId, uuid, time.Now())
	switch status {
	case DeleteConfirmFirstClick:
		_ = cb.Answer(c, 0, true, "Warning: Are you sure you want to delete this project? Click Delete again after 2 seconds to confirm.", "")
		return nil
	case DeleteConfirmTooSoon:
		_ = cb.Answer(c, 0, true, "Please wait 2 seconds before clicking Delete again.", "")
		return nil
	case DeleteConfirmExpired:
		_ = cb.Answer(c, 0, true, "Confirmation timed out. Click Delete again after 2 seconds to confirm.", "")
		return nil
	case DeleteConfirmProceed:
		_ = cb.Answer(c, 0, false, "Processing...", "")
	}

	err := config.Coolify.DeleteApplicationByUUID(uuid)
	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	if err != nil {
		_, err = cb.EditMessageText(c, "Delete failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	_, err = cb.EditMessageText(c, "Application deleted successfully.", &td.EditTextMessageOpts{ReplyMarkup: kb})
	return err
}

func scheduleMenuHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "sch_m:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔄 Restart",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("sch_a:" + uuid + ":restart"),
					},
				},
			},
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	_, err := cb.EditMessageText(c, "<b>📅 Select Action Type:</b>", &td.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}

func scheduleActionHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "Processing...", "")

	// Format: sch_a:uuid:actionType
	cbData := cb.DataString()
	data := strings.TrimPrefix(cbData, "sch_a:")
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return nil
	}
	uuid := parts[0]
	actionType := parts[1]

	// Common intervals
	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "Hourly",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_1h", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "Daily",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_1d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "Every 2 Days",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_2d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "Every 3 Days",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_3d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "Weekly",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_7d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("sch_m:" + uuid),
					},
				},
			},
		},
	}

	_, err := cb.EditMessageText(c, "<b>⏰ Select Schedule:</b>", &td.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}

func scheduleCreateHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "You are not authorized.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "Processing...", "")

	// Format: sch_c:uuid:actionType:schedule
	data := strings.TrimPrefix(cb.DataString(), "sch_c:")

	parts := strings.Split(data, ":")
	if len(parts) < 3 {
		return nil
	}
	uuid := parts[0]
	actionType := parts[1]
	schedule := parts[2]

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "Failed to get application: "+err.Error(), nil)
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
		_, _ = cb.EditMessageText(c, "Failed to save task: "+err.Error(), nil)
		return nil
	}

	if err := scheduler.ScheduleTask(task); err != nil {
		_ = database.DeleteTask(task.ID.Hex())
		_, _ = cb.EditMessageText(c, "Failed to schedule task: "+err.Error(), nil)
		return nil
	}

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	_, err = cb.EditMessageText(c, fmt.Sprintf("Task scheduled successfully!\n\nID: <code>%s</code>\nType: %s\nSchedule: %s", task.ID.Hex(), actionType, schedule), &td.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}
