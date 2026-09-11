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
				makeCallbackButton("App Tasks", "app_tasks:"+uuid),
				makeCallbackButton("Stop", "stop:"+uuid),
			},
			{
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

func deploymentsCallbackHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	kb := makeBackButton("start_menu")
	deps, err := config.Coolify.ListDeployments()
	if err != nil {
		return editCallback(c, cb, "Failed to fetch deployments: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	if len(deps) == 0 {
		return editCallback(c, cb, "No active deployments found.", &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	var sb strings.Builder
	sb.WriteString("<b>Active Deployments:</b>\n\n")
	for _, dep := range deps {
		sb.WriteString(fmt.Sprintf("ID: <code>%s</code>\n", dep.DeploymentUUID))
		sb.WriteString(fmt.Sprintf("Status: <code>%s</code>\n", dep.Status))
		if dep.Commit != "" {
			sb.WriteString(fmt.Sprintf("Commit: <code>%s</code>\n", dep.Commit))
		}
		sb.WriteString("--------------------\n")
	}

	return editCallback(c, cb, sb.String(), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serversCallbackHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	servers, err := config.Coolify.ListServers()
	if err != nil {
		return editCallback(c, cb, "Failed to fetch servers: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("start_menu")})
	}

	if len(servers) == 0 {
		return editCallback(c, cb, "No servers found.", &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("start_menu")})
	}

	kb := &td.ReplyMarkupInlineKeyboard{}
	for _, s := range servers {
		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			makeCallbackButton(s.Name, "srv_m:"+s.UUID),
		})
	}
	kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
		makeCallbackButton("Back", "start_menu"),
	})

	return editCallback(c, cb, "<b>Select a Server:</b>", &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serverMenuHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "srv_m:")
	servers, err := config.Coolify.ListServers()
	if err != nil {
		return editCallback(c, cb, "Failed to fetch server info: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("servers:")})
	}

	var s *coolify.Server
	for idx := range servers {
		if servers[idx].UUID == uuid {
			s = &servers[idx]
			break
		}
	}

	if s == nil {
		return editCallback(c, cb, "Server not found.", &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("servers:")})
	}

	text := fmt.Sprintf("<b>Server: %s</b>\nIP: <code>%s</code>\nUser: <code>%s</code>\nPort: %d", s.Name, s.IP, s.User, s.Port)
	if s.Description != "" {
		text += fmt.Sprintf("\nDescription: %s", s.Description)
	}

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				makeCallbackButton("Resources", "srv_res:"+uuid),
				makeCallbackButton("Domains", "srv_dom:"+uuid),
			},
			{
				makeCallbackButton("Validate", "srv_val:"+uuid),
				makeCallbackButton("Docker Cleanup", "srv_clean:"+uuid),
			},
			{
				makeCallbackButton("Restart Proxy", "srv_prx:"+uuid),
			},
			{
				makeCallbackButton("Back", "servers:"),
			},
		},
	}

	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serverResourcesHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "srv_res:")
	kb := makeBackButton("srv_m:" + uuid)

	resources, err := config.Coolify.GetServerResources(uuid)
	if err != nil {
		return editCallback(c, cb, "Failed to fetch resources: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	if len(resources) == 0 {
		return editCallback(c, cb, "No resources found for this server.", &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	var sb strings.Builder
	sb.WriteString("<b>Server Resources:</b>\n\n")
	for _, res := range resources {
		sb.WriteString(fmt.Sprintf("<b>%s</b> (%s)\nStatus: <code>%s</code>\nUUID: <code>%s</code>\n", res.Name, res.Type, res.Status, res.UUID))
		sb.WriteString("--------------------\n")
	}

	return editCallback(c, cb, sb.String(), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serverDomainsHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "srv_dom:")
	kb := makeBackButton("srv_m:" + uuid)

	domains, err := config.Coolify.GetServerDomains(uuid)
	if err != nil {
		return editCallback(c, cb, "Failed to fetch domains: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	if len(domains) == 0 {
		return editCallback(c, cb, "No domains found for this server.", &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	var sb strings.Builder
	sb.WriteString("<b>Server Domains:</b>\n\n")
	for _, d := range domains {
		sb.WriteString(fmt.Sprintf("IP: <code>%s</code>\nDomains: %s\n", d.IP, strings.Join(d.Domains, ", ")))
		sb.WriteString("--------------------\n")
	}

	return editCallback(c, cb, sb.String(), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serverValidateHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "srv_val:")
	kb := makeBackButton("srv_m:" + uuid)

	res, err := config.Coolify.ValidateServer(uuid)
	if err != nil {
		return editCallback(c, cb, "Validation failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	return editCallback(c, cb, fmt.Sprintf("Validation result: %s", res.Message), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serverCleanupHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "srv_clean:")
	kb := makeBackButton("srv_m:" + uuid)

	res, err := config.Coolify.RunDockerCleanup(uuid)
	if err != nil {
		return editCallback(c, cb, "Docker cleanup failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	return editCallback(c, cb, fmt.Sprintf("Cleanup result: %s", res.Message), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serverProxyRestartHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "srv_prx:")
	kb := makeBackButton("srv_m:" + uuid)

	res, err := config.Coolify.RestartServerProxy(uuid)
	if err != nil {
		return editCallback(c, cb, "Proxy restart failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	return editCallback(c, cb, fmt.Sprintf("Proxy restart result: %s", res.Message), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func databasesCallbackHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	dbs, err := config.Coolify.ListDatabases()
	if err != nil {
		return editCallback(c, cb, "Failed to fetch databases: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("start_menu")})
	}

	if len(dbs) == 0 {
		return editCallback(c, cb, "No databases found.", &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("start_menu")})
	}

	kb := &td.ReplyMarkupInlineKeyboard{}
	for _, db := range dbs {
		text := fmt.Sprintf("%s (%s)", db.Name, db.Type)
		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			makeCallbackButton(text, "db_m:"+db.UUID),
		})
	}
	kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
		makeCallbackButton("Back", "start_menu"),
	})

	return editCallback(c, cb, "<b>Select a Database:</b>", &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func databaseMenuHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "db_m:")
	db, err := config.Coolify.GetDatabaseByUUID(uuid)
	if err != nil {
		return editCallback(c, cb, "Failed to fetch database: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("databases:")})
	}

	text := fmt.Sprintf("<b>Database: %s</b>\nType: %s\nStatus: <code>%s</code>", db.Name, db.Type, db.Status)

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				makeCallbackButton("Start", "db_start:"+uuid),
				makeCallbackButton("Restart", "db_rest:"+uuid),
			},
			{
				makeCallbackButton("Logs", "db_logs:"+uuid),
				makeCallbackButton("Stop", "db_stop:"+uuid),
			},
			{
				makeCallbackButton("Back", "databases:"),
			},
		},
	}

	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func databaseStartHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "db_start:")
	kb := makeBackButton("db_m:" + uuid)

	res, err := config.Coolify.StartDatabaseByUUID(uuid)
	if err != nil {
		return editCallback(c, cb, "Database start failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	return editCallback(c, cb, fmt.Sprintf("Database start result: %s", res.Message), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func databaseStopHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "db_stop:")
	kb := makeBackButton("db_m:" + uuid)

	res, err := config.Coolify.StopDatabaseByUUID(uuid)
	if err != nil {
		return editCallback(c, cb, "Database stop failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	return editCallback(c, cb, fmt.Sprintf("Database stop result: %s", res.Message), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func databaseRestartHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "db_rest:")
	kb := makeBackButton("db_m:" + uuid)

	res, err := config.Coolify.RestartDatabaseByUUID(uuid)
	if err != nil {
		return editCallback(c, cb, "Database restart failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	return editCallback(c, cb, fmt.Sprintf("Database restart result: %s", res.Message), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func databaseLogsHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "db_logs:")
	kb := makeBackButton("db_m:" + uuid)

	_ = editCallback(c, cb, "Fetching database logs...", nil)
	logsData, err := config.Coolify.GetDatabaseLogsByUUID(uuid)
	if err != nil {
		_ = editCallback(c, cb, "Database logs error: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	db, _ := config.Coolify.GetDatabaseByUUID(uuid)
	dbName := uuid
	if db != nil {
		dbName = db.Name
	}

	if len(logsData) <= 3000 {
		text := fmt.Sprintf("<b>Logs for database %s:</b>\n\n<code>%s</code>", dbName, logsData)
		return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	tmpFile, err := os.CreateTemp("", "db-logs-*.txt")
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
		caption := fmt.Sprintf("<b>%s logs</b>", dbName)
		_, err = msg.ReplyDocument(c, td.InputFileLocal{Path: tmpFile.Name()}, docOpts(msg, caption, nil))
		if err == nil {
			_ = editCallback(c, cb, fmt.Sprintf("Logs for database <b>%s</b> sent as document.", dbName), &td.EditTextMessageOpts{ReplyMarkup: kb})
		}
	}
	return err
}

// Services handlers

func servicesCallbackHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	services, err := config.Coolify.ListServices()
	if err != nil {
		return editCallback(c, cb, "Failed to fetch services: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("start_menu")})
	}

	if len(services) == 0 {
		return editCallback(c, cb, "No services found.", &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("start_menu")})
	}

	kb := &td.ReplyMarkupInlineKeyboard{}
	for _, svc := range services {
		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			makeCallbackButton(svc.Name, "svc_m:"+svc.UUID),
		})
	}
	kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
		makeCallbackButton("Back", "start_menu"),
	})

	return editCallback(c, cb, "<b>Select a Service:</b>", &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serviceMenuHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "svc_m:")
	svc, err := config.Coolify.GetServiceByUUID(uuid)
	if err != nil {
		return editCallback(c, cb, "Failed to fetch service: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("services:")})
	}

	text := fmt.Sprintf("<b>Service: %s</b>\nType: <code>%s</code>", svc.Name, svc.ServiceType)
	if svc.Description != "" {
		text += fmt.Sprintf("\nDescription: %s", svc.Description)
	}

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				makeCallbackButton("Start", "svc_start:"+uuid),
				makeCallbackButton("Restart", "svc_rest:"+uuid),
			},
			{
				makeCallbackButton("Logs", "svc_logs:"+uuid),
				makeCallbackButton("ENV", "svc_env:"+uuid),
			},
			{
				makeCallbackButton("Stop", "svc_stop:"+uuid),
			},
			{
				makeCallbackButton("Back", "services:"),
			},
		},
	}

	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serviceStartHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "svc_start:")
	kb := makeBackButton("svc_m:" + uuid)

	res, err := config.Coolify.StartServiceByUUID(uuid)
	if err != nil {
		return editCallback(c, cb, "Service start failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	return editCallback(c, cb, fmt.Sprintf("Service start result: %s", res.Message), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serviceStopHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "svc_stop:")
	kb := makeBackButton("svc_m:" + uuid)

	res, err := config.Coolify.StopServiceByUUID(uuid)
	if err != nil {
		return editCallback(c, cb, "Service stop failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	return editCallback(c, cb, fmt.Sprintf("Service stop result: %s", res.Message), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serviceRestartHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "svc_rest:")
	kb := makeBackButton("svc_m:" + uuid)

	res, err := config.Coolify.RestartServiceByUUID(uuid)
	if err != nil {
		return editCallback(c, cb, "Service restart failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	return editCallback(c, cb, fmt.Sprintf("Service restart result: %s", res.Message), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func serviceLogsHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "svc_logs:")
	kb := makeBackButton("svc_m:" + uuid)

	_ = editCallback(c, cb, "Fetching service logs...", nil)
	logsData, err := config.Coolify.GetServiceLogsByUUID(uuid)
	if err != nil {
		_ = editCallback(c, cb, "Service logs error: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	svc, _ := config.Coolify.GetServiceByUUID(uuid)
	svcName := uuid
	if svc != nil {
		svcName = svc.Name
	}

	if len(logsData) <= 3000 {
		text := fmt.Sprintf("<b>Logs for service %s:</b>\n\n<code>%s</code>", svcName, logsData)
		return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	tmpFile, err := os.CreateTemp("", "svc-logs-*.txt")
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
		caption := fmt.Sprintf("<b>%s logs</b>", svcName)
		_, err = msg.ReplyDocument(c, td.InputFileLocal{Path: tmpFile.Name()}, docOpts(msg, caption, nil))
		if err == nil {
			_ = editCallback(c, cb, fmt.Sprintf("Logs for service <b>%s</b> sent as document.", svcName), &td.EditTextMessageOpts{ReplyMarkup: kb})
		}
	}
	return err
}

func serviceEnvHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !cb.IsPrivate() {
		_ = cb.Answer(c, 0, true, "Environment variables can only be viewed in private chat.", "")
		return nil
	}

	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "svc_env:")
	svc, err := config.Coolify.GetServiceByUUID(uuid)
	if err != nil {
		_ = editCallback(c, cb, "Failed to load service: "+err.Error(), nil)
		return nil
	}

	envs, err := config.Coolify.GetServiceEnvsByUUID(uuid)
	kb := makeBackButton("svc_m:" + uuid)

	if err != nil {
		_ = editCallback(c, cb, fmt.Sprintf("Failed to fetch service envs: %v", err), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	if len(envs) == 0 {
		_ = editCallback(c, cb, fmt.Sprintf("No environment variables found for service <b>%s</b>.", svc.Name), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Environment Variables for Service %s</b>\n\n", svc.Name))
	for _, env := range envs {
		val := env.Value
		if val == "" {
			val = env.RealValue
		}
		sb.WriteString(fmt.Sprintf("<code>%s</code> = <code>%s</code>\n", env.Key, val))
	}

	return editCallback(c, cb, sb.String(), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

// App tasks, Team, Tags handlers

func appTasksHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "app_tasks:")
	tasks, err := config.Coolify.ListApplicationScheduledTasks(uuid)
	kb := makeBackButton("project_menu:" + uuid)

	if err != nil {
		return editCallback(c, cb, "Failed to fetch scheduled tasks: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	if len(tasks) == 0 {
		return editCallback(c, cb, "No scheduled tasks found for this application.", &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	kbTasks := &td.ReplyMarkupInlineKeyboard{}
	var sb strings.Builder
	sb.WriteString("<b>Application Scheduled Tasks:</b>\n\n")

	for _, task := range tasks {
		sb.WriteString(fmt.Sprintf("<b>%s</b>\nCommand: <code>%s</code>\nFrequency: <code>%s</code>\n", task.Name, task.Command, task.Frequency))
		sb.WriteString("--------------------\n")

		btnText := fmt.Sprintf("Run: %s", task.Name)
		cbData := fmt.Sprintf("app_exec:%s:%s", uuid, task.UUID)
		kbTasks.Rows = append(kbTasks.Rows, []td.InlineKeyboardButton{
			makeCallbackButton(btnText, cbData),
		})
	}
	kbTasks.Rows = append(kbTasks.Rows, []td.InlineKeyboardButton{
		makeCallbackButton("Back", "project_menu:"+uuid),
	})

	return editCallback(c, cb, sb.String(), &td.EditTextMessageOpts{ReplyMarkup: kbTasks})
}

func appTaskExecuteHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	data := strings.TrimPrefix(cb.DataString(), "app_exec:")
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return nil
	}
	uuid, taskUUID := parts[0], parts[1]
	kb := makeBackButton("app_tasks:" + uuid)

	res, err := config.Coolify.ExecuteApplicationScheduledTask(uuid, taskUUID)
	if err != nil {
		return editCallback(c, cb, "Task execution failed: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	msg := "Task execution triggered successfully."
	if res.Message != "" {
		msg = res.Message
	}
	return editCallback(c, cb, msg, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func teamInfoHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	team, err := config.Coolify.GetTeam()
	kb := makeBackButton("start_menu")

	if err != nil {
		return editCallback(c, cb, "Failed to fetch team info: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	members, _ := config.Coolify.ListTeamMembers()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Team: %s</b>\n", team.Name))
	if team.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", team.Description))
	}
	sb.WriteString(fmt.Sprintf("Personal Team: <code>%t</code>\n\n", team.PersonalTeam))

	if len(members) > 0 {
		sb.WriteString("<b>Members:</b>\n")
		for _, m := range members {
			sb.WriteString(fmt.Sprintf("- <b>%s</b> (<code>%s</code>)\n", m.Name, m.Email))
		}
	} else if len(team.Members) > 0 {
		sb.WriteString("<b>Members:</b>\n")
		for _, m := range team.Members {
			sb.WriteString(fmt.Sprintf("- <b>%s</b> (<code>%s</code>)\n", m.Name, m.Email))
		}
	}

	return editCallback(c, cb, sb.String(), &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func tagsHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	tags, err := config.Coolify.ListTags()
	kb := makeBackButton("start_menu")

	if err != nil {
		return editCallback(c, cb, "Failed to fetch tags: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	if len(tags) == 0 {
		return editCallback(c, cb, "No tags found.", &td.EditTextMessageOpts{ReplyMarkup: kb})
	}

	var sb strings.Builder
	sb.WriteString("<b>Tags:</b>\n\n")
	for _, t := range tags {
		sb.WriteString(fmt.Sprintf("- <b>%s</b> (<code>%s</code>)\n", t.Name, t.UUID))
	}

	return editCallback(c, cb, sb.String(), &td.EditTextMessageOpts{ReplyMarkup: kb})
}
