package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"os"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func listProjectsHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("Processing...")
	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_, _ = cb.Edit("❌ Failed to fetch projects:" + err.Error())
		return nil
	}

	if len(apps) == 0 {
		_, _ = cb.Edit("😶 No applications found.")
		return nil
	}

	page := 1
	data := cb.DataString()
	if strings.Contains(data, ":") {
		parts := strings.Split(data, ":")
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &page)
		}
	}

	start, end, paginationButtons := Paginate(len(apps), page, 7, "list_projects:")
	kb := telegram.NewKeyboard()
	for _, app := range apps[start:end] {
		text := fmt.Sprintf("📦 %s (%s)", app.Name, app.Status)
		data := "project_menu:" + app.UUID
		kb.AddRow(telegram.Button.Data(text, data))
	}

	if len(paginationButtons) > 0 {
		var row []telegram.KeyboardButton
		for _, btn := range paginationButtons {
			row = append(row, telegram.Button.Data(btn.Text, btn.Data))
		}
		kb.AddRow(row...)
	}

	_, err = cb.Edit("<b>📋 Select a project:</b>", &telegram.SendOptions{ReplyMarkup: kb.Build()})
	return err
}

func projectMenuHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}

	_, _ = cb.Answer("Processing...")
	uuid := strings.TrimPrefix(cb.DataString(), "project_menu:")

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, err = cb.Edit("❌ Failed to load project: "+err.Error(), nil)
		return err
	}

	text := fmt.Sprintf("<b>📦 %s</b>\n🌐 %s\n📄 Status: <code>%s</code>", app.Name, app.FQDN, app.Status)
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔄 Restart", "restart:"+uuid), telegram.Button.Data("🚀 Deploy", "deploy:"+uuid)).
		AddRow(telegram.Button.Data("📜 Logs", "logs:"+uuid), telegram.Button.Data("ℹ️ Status", "status:"+uuid)).
		AddRow(telegram.Button.Data("📅 Schedule", "schedule_menu:"+uuid)).
		AddRow(telegram.Button.Data("🛑 Stop", "stop:"+uuid), telegram.Button.Data("❌ Delete", "delete:"+uuid)).
		AddRow(telegram.Button.Data("🔙 Back", "list_projects:"))

	_, err = cb.Edit(text, &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: keyboard.Build(),
	})
	return err
}

func restartHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("Processing...")
	uuid := strings.TrimPrefix(cb.DataString(), "restart:")

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 Back", "project_menu:"+uuid))

	res, err := config.Coolify.RestartApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.Edit("❌ Restart failed: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return nil
	}

	text := fmt.Sprintf("✅ Restart queued!\nDeployment UUID: <code>%s</code>", res.DeploymentUUID)
	_, err = cb.Edit(text, &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: keyboard.Build()})
	return err
}

func deployHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("Processing...")
	uuid := strings.TrimPrefix(cb.DataString(), "deploy:")
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 Back", "project_menu:"+uuid))
	res, err := config.Coolify.StartApplicationDeployment(uuid, false, false)
	if err != nil {
		_, _ = cb.Edit("❌ Deploy failed: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return err
	}
	text := fmt.Sprintf("✅ Deployment queued!\nDeployment UUID: <code>%s</code>", res.DeploymentUUID)
	_, err = cb.Edit(text, &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: keyboard.Build()})
	return err
}

func logsHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("Processing...")
	uuid := strings.TrimPrefix(cb.DataString(), "logs:")
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 Back", "project_menu:"+uuid))

	msg, _ := cb.Edit("Processing...")
	logsData, err := config.Coolify.GetApplicationLogsByUUID(uuid)
	if err != nil {
		_, _ = cb.Edit("❌ Logs error: "+err.Error(), &telegram.SendOptions{
			ReplyMarkup: keyboard.Build(),
		})
		return nil
	}

	tmpFile, err := os.CreateTemp("", "logs-*.txt")
	if err != nil {
		_, _ = cb.Edit("❌ Failed to create temp file: "+err.Error(), nil)
		return err
	}

	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(logsData)); err != nil {
		_, _ = cb.Edit("❌ Failed to write logs: "+err.Error(), nil)
		return err
	}
	tmpFile.Close()

	opts := telegram.SendOptions{
		Upload: &telegram.UploadOptions{
			ProgressCallback: func(pi *telegram.ProgressInfo) {
				msg.Edit(fmt.Sprintf("Uploading... %.2f%% complete (%.2f MB/s), ETA: %.2f seconds",
					pi.Percentage,
					pi.CurrentSpeed/1024/1024,
					pi.ETA,
				))
			},
			ProgressInterval: 5,
		},
		Media: tmpFile.Name(),
		Attributes: []telegram.DocumentAttribute{
			&telegram.DocumentAttributeFilename{
				FileName: tmpFile.Name(),
			},
		},
		Caption:     "LOGS",
		ReplyMarkup: keyboard.Build(),
	}
	_, err = msg.Edit("LOGS", &opts)
	if err != nil {
		_, _ = cb.Edit("❌ Failed to send logs: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return err
	}

	return nil
}

func statusHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("Processing...")
	uuid := strings.TrimPrefix(cb.DataString(), "status:")
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 Back", "project_menu:"+uuid))
	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.Edit("❌ Status error: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return nil
	}

	text := fmt.Sprintf("📦 <b>%s</b>\nCurrent Status: <code>%s</code>", app.Name, app.Status)
	_, err = cb.Edit(text, &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: keyboard.Build()})
	return err
}

func stopHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("Processing...")
	uuid := strings.TrimPrefix(cb.DataString(), "stop:")
	res, err := config.Coolify.StopApplicationByUUID(uuid)
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 Back", "project_menu:"+uuid))
	if err != nil {
		_, _ = cb.Edit("❌ Stop failed: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return nil
	}

	_, err = cb.Edit("🛑 "+res.Message, &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
	return err
}

func deleteHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}

	_, _ = cb.Answer("Processing...")
	uuid := strings.TrimPrefix(cb.DataString(), "delete:")
	err := config.Coolify.DeleteApplicationByUUID(uuid)
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 Back", "project_menu:"+uuid))
	if err != nil {
		_, err = cb.Edit("❌ Delete failed: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return nil
	}

	_, err = cb.Edit("✅ Application deleted successfully.", &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
	return err
}

func scheduleMenuHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("Processing...")
	uuid := strings.TrimPrefix(cb.DataString(), "schedule_menu:")

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔄 Restart", "schedule_action:restart:"+uuid)).
		AddRow(telegram.Button.Data("🔙 Back", "project_menu:"+uuid))

	_, err := cb.Edit("<b>📅 Select Action Type:</b>", &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: keyboard.Build(),
	})
	return err
}

func scheduleActionHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("Processing...")
	// Format: schedule_action:restart:uuid
	data := strings.TrimPrefix(cb.DataString(), "schedule_action:")
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return nil
	}
	actionType := parts[0]
	uuid := parts[1]

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("Hourly", fmt.Sprintf("schedule_create:%s:%s:every_1h", actionType, uuid))).
		AddRow(telegram.Button.Data("Daily", fmt.Sprintf("schedule_create:%s:%s:every_1d", actionType, uuid))).
		AddRow(telegram.Button.Data("Every 2 Days", fmt.Sprintf("schedule_create:%s:%s:every_2d", actionType, uuid))).
		AddRow(telegram.Button.Data("Every 3 Days", fmt.Sprintf("schedule_create:%s:%s:every_3d", actionType, uuid))).
		AddRow(telegram.Button.Data("Weekly", fmt.Sprintf("schedule_create:%s:%s:every_7d", actionType, uuid))).
		AddRow(telegram.Button.Data("🔙 Back", "schedule_menu:"+uuid))

	_, err := cb.Edit("<b>⏰ Select Schedule:</b>", &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: keyboard.Build(),
	})
	return err
}

func scheduleCreateHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 You are not authorized.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("Scheduling...")
	// Format: schedule_create:restart:uuid:every_1h
	data := strings.TrimPrefix(cb.DataString(), "schedule_create:")
	parts := strings.Split(data, ":")
	if len(parts) < 3 {
		return nil
	}
	actionType := parts[0]
	uuid := parts[1]
	schedule := parts[2]

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.Edit("❌ Failed to get application: " + err.Error())
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
		_, _ = cb.Edit("❌ Failed to save task: " + err.Error())
		return nil
	}

	if err := scheduler.ScheduleTask(task); err != nil {
		_ = database.DeleteTask(task.ID.Hex())
		_, _ = cb.Edit("❌ Failed to schedule task: " + err.Error())
		return nil
	}

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 Back", "project_menu:"+uuid))

	_, err = cb.Edit(fmt.Sprintf("✅ Task scheduled successfully!\n\nID: <code>%s</code>\nType: %s\nSchedule: %s", task.ID.Hex(), actionType, schedule), &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: keyboard.Build(),
	})
	return err
}
