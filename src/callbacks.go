package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/coolity"
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
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	if cbData == "list_projects:projects" {
		projects, err := config.Coolify.ListProjects()
		if err != nil || len(projects) == 0 {
			return editCallback(c, cb, "No projects found.", &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("list_projects:")})
		}

		kb := &td.ReplyMarkupInlineKeyboard{}
		for _, proj := range projects {
			kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
				makeCallbackButton(proj.Name, "proj:"+proj.UUID),
			})
		}
		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			makeCallbackButton("All Applications", "list_projects:"),
		})
		return editCallback(c, cb, "<b>Select a Project:</b>", &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_ = editCallback(c, cb, "Failed to fetch applications: "+err.Error(), nil)
		return nil
	}

	if len(apps) == 0 {
		_ = editCallback(c, cb, "No applications found.", nil)
		return nil
	}

	page := 1
	if strings.Contains(cbData, ":") {
		parts := strings.Split(cbData, ":")
		if len(parts) > 1 && parts[1] != "" && parts[1] != "projects" {
			fmt.Sscanf(parts[1], "%d", &page)
		}
	}

	start, end, paginationButtons := Paginate(len(apps), page, 7, "list_projects:")

	kb := &td.ReplyMarkupInlineKeyboard{}
	for _, app := range apps[start:end] {
		text := fmt.Sprintf("%s (%s)", app.Name, app.Status)
		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			makeCallbackButton(text, "project_menu:"+app.UUID),
		})
	}

	row := buildPaginationButtonsRow(paginationButtons)
	row = append(row, makeCallbackButton("Projects", "list_projects:projects"))
	kb.Rows = append(kb.Rows, row)

	return editCallback(c, cb, "<b>Select an Application:</b>", &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func projectSelectHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	raw := strings.TrimPrefix(cbData, "proj:")
	parts := strings.Split(raw, ":")

	projectUUID := parts[0]
	page := 1
	if len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &page)
	}

	project, _ := config.Coolify.GetProjectByUUID(projectUUID)
	projectName := "Project"
	if project != nil {
		projectName = project.Name
	}

	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_ = editCallback(c, cb, "Failed to fetch applications: "+err.Error(), nil)
		return nil
	}

	var projectApps []coolify.Application
	for _, app := range apps {
		if projectUUID == "" || app.ProjectUUID == "" || app.ProjectUUID == projectUUID {
			projectApps = append(projectApps, app)
		}
	}
	if len(projectApps) == 0 {
		projectApps = apps
	}

	prefix := fmt.Sprintf("proj:%s:", projectUUID)
	start, end, paginationButtons := Paginate(len(projectApps), page, 7, prefix)

	kb := &td.ReplyMarkupInlineKeyboard{}
	for _, app := range projectApps[start:end] {
		text := fmt.Sprintf("%s (%s)", app.Name, app.Status)
		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			makeCallbackButton(text, "project_menu:"+app.UUID),
		})
	}

	if row := buildPaginationButtonsRow(paginationButtons); len(row) > 0 {
		kb.Rows = append(kb.Rows, row)
	}

	kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
		makeCallbackButton("Back", "list_projects:projects"),
		makeCallbackButton("All Applications", "list_projects:"),
	})

	return editCallback(c, cb, fmt.Sprintf("<b>Applications in %s:</b>", projectName), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func projectMenuHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "project_menu:")
	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		return editCallback(c, cb, "Failed to load project: "+err.Error(), nil)
	}

	text := fmt.Sprintf("<b>%s</b>\nURL: %s\nStatus: <code>%s</code>", app.Name, app.FQDN, app.Status)

	backData := "list_projects:"
	if app.ProjectUUID != "" {
		backData = "proj:" + app.ProjectUUID
	}

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				makeCallbackButton("Restart", "restart:"+uuid),
				makeCallbackButton("Deploy", "deploy:"+uuid),
			},
			{
				makeCallbackButton("Logs", "logs:"+uuid),
				makeCallbackButton("Status", "status:"+uuid),
			},
			{
				makeCallbackButton("ENV", "env:"+uuid),
				makeCallbackButton("Schedule", "sch_m:"+uuid),
			},
			{
				makeCallbackButton("Stop", "stop:"+uuid),
				makeCallbackButton("Delete", "delete:"+uuid),
			},
			{
				makeCallbackButton("Back", backData),
				makeCallbackButton("Projects", "list_projects:projects"),
			},
		},
	}

	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func envCallbackHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !cb.IsPrivate() {
		_ = cb.Answer(c, 0, true, "Environment variables can only be viewed in private chat.", "")
		return nil
	}

	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "env:")

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_ = editCallback(c, cb, "Failed to load project: "+err.Error(), nil)
		return nil
	}

	envs, err := config.Coolify.GetApplicationEnvsByUUID(uuid)
	kb := makeBackButton("project_menu:" + uuid)

	if err != nil {
		_ = editCallback(c, cb, fmt.Sprintf("Failed to fetch environment variables: %v", err), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	if len(envs) == 0 {
		_ = editCallback(c, cb, fmt.Sprintf("No environment variables found for <b>%s</b>.", app.Name), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Environment Variables for %s</b>\n\n", app.Name))
	for _, env := range envs {
		val := env.Value
		if val == "" {
			val = env.RealValue
		}
		sb.WriteString(fmt.Sprintf("<code>%s</code> = <code>%s</code>\n", env.Key, val))
	}

	return editCallback(c, cb, sb.String(), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func restartHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "restart:")
	kb := makeBackButton("project_menu:" + uuid)

	res, err := config.Coolify.RestartApplicationByUUID(uuid)
	if err != nil {
		_ = editCallback(c, cb, "Restart failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	text := "Restart queued!"
	if res.DeploymentUUID != "" {
		text += fmt.Sprintf("\nDeployment UUID: <code>%s</code>", res.DeploymentUUID)
	} else if res.Message != "" {
		text += fmt.Sprintf("\nMessage: %s", res.Message)
	}
	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func deployHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "deploy:")
	kb := makeBackButton("project_menu:" + uuid)

	res, err := config.Coolify.StartApplicationDeployment(uuid, false, false)
	if err != nil {
		_ = editCallback(c, cb, "Deploy failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return err
	}

	text := fmt.Sprintf("Deployment queued!\nDeployment UUID: <code>%s</code>", res.DeploymentUUID)
	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func logsHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "logs:")
	kb := makeBackButton("project_menu:" + uuid)

	_ = editCallback(c, cb, "Fetching logs...", nil)
	logsData, err := config.Coolify.GetApplicationLogsByUUID(uuid)
	if err != nil {
		_ = editCallback(c, cb, "Logs error: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	app, _ := config.Coolify.GetApplicationByUUID(uuid)
	appName := uuid
	if app != nil {
		appName = app.Name
	}

	if len(logsData) <= 3000 {
		text := fmt.Sprintf("<b>Logs for %s:</b>\n\n<code>%s</code>", appName, logsData)
		return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	tmpFile, err := os.CreateTemp("", "logs-*.txt")
	if err != nil {
		_ = editCallback(c, cb, "Failed to create temp file: "+err.Error(), nil)
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write([]byte(logsData)); err != nil {
		_ = editCallback(c, cb, "Failed to write logs: "+err.Error(), nil)
		return err
	}
	tmpFile.Close()

	msg, _ := cb.GetMessage(c)
	if msg != nil {
		caption := fmt.Sprintf("<b>%s logs</b>", appName)
		_, err = msg.ReplyDocument(c, td.InputFileLocal{Path: tmpFile.Name()}, docOpts(msg, caption, nil))
		if err == nil {
			_ = editCallback(c, cb, fmt.Sprintf("Logs for <b>%s</b> sent as document.", appName), &td.EditTextMessageOpts{ReplyMarkup: kb})
		}
	}
	return err
}

func statusHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, true, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "status:")
	kb := makeBackButton("project_menu:" + uuid)

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_ = editCallback(c, cb, "Status error: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	text := fmt.Sprintf("<b>%s</b>\nCurrent Status: <code>%s</code>", app.Name, app.Status)
	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func stopHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "stop:")
	kb := makeBackButton("project_menu:" + uuid)

	res, err := config.Coolify.StopApplicationByUUID(uuid)
	if err != nil {
		_ = editCallback(c, cb, "Stop failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	return editCallback(c, cb, res.Message, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func deleteHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
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
	kb := makeBackButton("project_menu:" + uuid)

	if err != nil {
		_ = editCallback(c, cb, "Delete failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	return editCallback(c, cb, "Application deleted successfully.", &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func scheduleMenuHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "sch_m:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				makeCallbackButton("Restart", "sch_a:"+uuid+":restart"),
			},
			{
				makeCallbackButton("Back", "project_menu:"+uuid),
			},
		},
	}

	return editCallback(c, cb, "<b>Select Action Type:</b>", &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func scheduleActionHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	cbData := cb.DataString()
	data := strings.TrimPrefix(cbData, "sch_a:")
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return nil
	}
	uuid := parts[0]
	actionType := parts[1]

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{makeCallbackButton("Hourly", fmt.Sprintf("sch_c:%s:%s:every_1h", uuid, actionType))},
			{makeCallbackButton("Daily", fmt.Sprintf("sch_c:%s:%s:every_1d", uuid, actionType))},
			{makeCallbackButton("Every 2 Days", fmt.Sprintf("sch_c:%s:%s:every_2d", uuid, actionType))},
			{makeCallbackButton("Every 3 Days", fmt.Sprintf("sch_c:%s:%s:every_3d", uuid, actionType))},
			{makeCallbackButton("Weekly", fmt.Sprintf("sch_c:%s:%s:every_7d", uuid, actionType))},
			{makeCallbackButton("Back", "sch_m:"+uuid)},
		},
	}

	return editCallback(c, cb, "<b>Select Schedule:</b>", &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func scheduleCreateHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

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
		_ = editCallback(c, cb, "Failed to get application: "+err.Error(), nil)
		return nil
	}

	task := database.ScheduledTask{
		ID:          bson.NewObjectID(),
		Name:        app.Name,
		ProjectUUID: uuid,
		Type:        actionType,
		Schedule:    schedule,
	}

	if err = database.AddTask(task); err != nil {
		_ = editCallback(c, cb, "Failed to save task: "+err.Error(), nil)
		return nil
	}

	if err = scheduler.ScheduleTask(task); err != nil {
		_ = database.DeleteTask(task.ID.Hex())
		_ = editCallback(c, cb, "Failed to schedule task: "+err.Error(), nil)
		return nil
	}

	kb := makeBackButton("project_menu:" + uuid)
	return editCallback(c, cb, fmt.Sprintf("Task scheduled successfully!\n\nID: <code>%s</code>\nType: %s\nSchedule: %s", task.ID.Hex(), actionType, schedule), &td.EditTextMessageOpts{ReplyMarkup: kb})
}
