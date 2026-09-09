package src

import (
	"fmt"
	"os"
	"strings"

	"coolifymanager/src/config"
	"coolifymanager/src/coolity"

	td "github.com/AshokShau/gotdbot"
)

func findApplication(query string) (*coolify.Application, error) {
	apps, err := config.Coolify.ListApplications()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch applications: %w", err)
	}

	for _, app := range apps {
		if strings.EqualFold(app.UUID, query) || strings.EqualFold(app.Name, query) {
			return &app, nil
		}
	}

	return nil, fmt.Errorf("application '%s' not found", query)
}

func promptAppSelection(c *td.Client, msg *td.Message, actionPrefix, title, usage string) error {
	apps, err := config.Coolify.ListApplications()
	if err != nil || len(apps) == 0 {
		_, err = msg.ReplyText(c, usage, sendOpts(msg, nil))
		return err
	}

	kb := &td.ReplyMarkupInlineKeyboard{}
	for _, app := range apps {
		label := app.Name
		if actionPrefix == "deploy:" {
			label = "Deploy " + app.Name
		}
		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			makeCallbackButton(label, actionPrefix+app.UUID),
		})
	}
	_, err = msg.ReplyText(c, title, sendOpts(msg, kb))
	return err
}

func appsHandler(c *td.Client, msg *td.Message) error {
	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_, err = msg.ReplyText(c, "Failed to fetch projects: "+err.Error(), sendOpts(msg, nil))
		return err
	}

	if len(apps) == 0 {
		_, err = msg.ReplyText(c, "No applications found.", sendOpts(msg, nil))
		return err
	}

	start, end, paginationButtons := Paginate(len(apps), 1, 7, "list_projects:")

	kb := &td.ReplyMarkupInlineKeyboard{}
	for _, app := range apps[start:end] {
		text := fmt.Sprintf("%s (%s)", app.Name, app.Status)
		data := "project_menu:" + app.UUID
		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			makeCallbackButton(text, data),
		})
	}

	if row := buildPaginationButtonsRow(paginationButtons); len(row) > 0 {
		kb.Rows = append(kb.Rows, row)
	}

	_, err = msg.ReplyText(c, "<b>Applications List:</b>", sendOpts(msg, kb))
	return err
}

func envCmdHandler(c *td.Client, msg *td.Message) error {
	args := strings.Fields(msg.Text())
	if len(args) < 2 {
		return promptAppSelection(c, msg, "env:", "<b>Select an application to view ENV:</b>", "Usage: /env <app_name_or_uuid>")
	}

	target := args[1]
	app, err := findApplication(target)
	if err != nil {
		_, err = msg.ReplyText(c, err.Error(), sendOpts(msg, nil))
		return err
	}

	envs, err := config.Coolify.GetApplicationEnvsByUUID(app.UUID)
	if err != nil {
		_, err = msg.ReplyText(c, fmt.Sprintf("Failed to fetch environment variables for %s: %v", app.Name, err), sendOpts(msg, nil))
		return err
	}

	if len(envs) == 0 {
		_, err = msg.ReplyText(c, fmt.Sprintf("No environment variables found for <b>%s</b>.", app.Name), sendOpts(msg, nil))
		return err
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

	kb := makeBackButton("project_menu:" + app.UUID)
	_, err = msg.ReplyText(c, sb.String(), sendOpts(msg, kb))
	return err
}

func logsCmdHandler(c *td.Client, msg *td.Message) error {
	args := strings.Fields(msg.Text())
	if len(args) < 2 {
		return promptAppSelection(c, msg, "logs:", "<b>Select an application to view logs:</b>", "Usage: /logs <app_name_or_uuid>")
	}

	target := args[1]
	app, err := findApplication(target)
	if err != nil {
		_, err = msg.ReplyText(c, err.Error(), sendOpts(msg, nil))
		return err
	}

	logsData, err := config.Coolify.GetApplicationLogsByUUID(app.UUID)
	if err != nil {
		_, err = msg.ReplyText(c, fmt.Sprintf("Failed to fetch logs for %s: %v", app.Name, err), sendOpts(msg, nil))
		return err
	}

	kb := makeBackButton("project_menu:" + app.UUID)

	if len(logsData) <= 3000 {
		text := fmt.Sprintf("<b>%s logs:</b>\n\n<code>%s</code>", app.Name, logsData)
		_, err = msg.ReplyText(c, text, sendOpts(msg, kb))
		return err
	}

	tmpFile, err := os.CreateTemp("", "logs-*.txt")
	if err != nil {
		_, err = msg.ReplyText(c, "Failed to create temp logs file: "+err.Error(), sendOpts(msg, nil))
		return err
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.Write([]byte(logsData))
	tmpFile.Close()

	caption := fmt.Sprintf("<b>%s logs</b>", app.Name)
	_, err = msg.ReplyDocument(c, td.InputFileLocal{Path: tmpFile.Name()}, docOpts(msg, caption, kb))
	return err
}

func statusCmdHandler(c *td.Client, msg *td.Message) error {
	args := strings.Fields(msg.Text())
	if len(args) < 2 {
		return appsHandler(c, msg)
	}

	target := args[1]
	app, err := findApplication(target)
	if err != nil {
		_, err = msg.ReplyText(c, err.Error(), sendOpts(msg, nil))
		return err
	}

	detail, err := config.Coolify.GetApplicationByUUID(app.UUID)
	if err != nil {
		_, err = msg.ReplyText(c, fmt.Sprintf("Failed to fetch status for %s: %v", app.Name, err), sendOpts(msg, nil))
		return err
	}

	text := fmt.Sprintf("<b>%s</b>\nURL: %s\nStatus: <code>%s</code>", detail.Name, detail.FQDN, detail.Status)
	if detail.Description != "" {
		text += fmt.Sprintf("\nDescription: %s", detail.Description)
	}
	if detail.GitRepository != "" {
		text += fmt.Sprintf("\nGit Repo: %s (%s)", detail.GitRepository, detail.GitBranch)
	}

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				makeCallbackButton("Restart", "restart:"+detail.UUID),
				makeCallbackButton("Deploy", "deploy:"+detail.UUID),
			},
			{
				makeCallbackButton("Back", "project_menu:"+detail.UUID),
			},
		},
	}

	_, err = msg.ReplyText(c, text, sendOpts(msg, kb))
	return err
}

func deployCmdHandler(c *td.Client, msg *td.Message) error {
	args := strings.Fields(msg.Text())
	if len(args) < 2 {
		return promptAppSelection(c, msg, "deploy:", "<b>Select an application to deploy:</b>", "Usage: /deploy <app_name_or_uuid>")
	}

	target := args[1]
	app, err := findApplication(target)
	if err != nil {
		_, err = msg.ReplyText(c, err.Error(), sendOpts(msg, nil))
		return err
	}

	loadingMsg, _ := msg.ReplyText(c, fmt.Sprintf("Deploying <b>%s</b>...", app.Name), sendOpts(msg, nil))

	res, err := config.Coolify.StartApplicationDeployment(app.UUID, false, false)
	kb := makeBackButton("project_menu:" + app.UUID)

	if err != nil {
		text := fmt.Sprintf("Deployment failed for %s: %v", app.Name, err)
		if loadingMsg != nil {
			_, _ = editMsg(c, loadingMsg, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
		} else {
			_, _ = msg.ReplyText(c, text, sendOpts(msg, kb))
		}
		return nil
	}

	resultText := fmt.Sprintf("Deployment started for <b>%s</b>\nDeployment UUID: <code>%s</code>", app.Name, res.DeploymentUUID)
	if loadingMsg != nil {
		_, err = editMsg(c, loadingMsg, resultText, &td.EditTextMessageOpts{ReplyMarkup: kb})
	} else {
		_, err = msg.ReplyText(c, resultText, sendOpts(msg, kb))
	}
	return err
}

func deploymentsHandler(c *td.Client, msg *td.Message) error {
	deps, err := config.Coolify.ListDeployments()
	if err != nil {
		_, err = msg.ReplyText(c, "Failed to fetch deployments: "+err.Error(), sendOpts(msg, nil))
		return err
	}

	if len(deps) == 0 {
		_, err = msg.ReplyText(c, "No active deployments found.", sendOpts(msg, nil))
		return err
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

	_, err = msg.ReplyText(c, sb.String(), sendOpts(msg, nil))
	return err
}

func serversHandler(c *td.Client, msg *td.Message) error {
	servers, err := config.Coolify.ListServers()
	if err != nil {
		_, err = msg.ReplyText(c, "Failed to fetch servers: "+err.Error(), sendOpts(msg, nil))
		return err
	}

	if len(servers) == 0 {
		_, err = msg.ReplyText(c, "No servers found.", sendOpts(msg, nil))
		return err
	}

	var sb strings.Builder
	sb.WriteString("<b>Registered Servers:</b>\n\n")
	for _, server := range servers {
		sb.WriteString(fmt.Sprintf("<b>%s</b>\n", server.Name))
		if server.IP != "" {
			sb.WriteString(fmt.Sprintf("IP: <code>%s</code>\n", server.IP))
		}
		if server.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", server.Description))
		}
		sb.WriteString("--------------------\n")
	}

	_, err = msg.ReplyText(c, sb.String(), sendOpts(msg, nil))
	return err
}

func databasesHandler(c *td.Client, msg *td.Message) error {
	dbs, err := config.Coolify.ListDatabases()
	if err != nil {
		_, err = msg.ReplyText(c, "Failed to fetch databases: "+err.Error(), sendOpts(msg, nil))
		return err
	}

	if len(dbs) == 0 {
		_, err = msg.ReplyText(c, "No databases found.", sendOpts(msg, nil))
		return err
	}

	var sb strings.Builder
	sb.WriteString("<b>Registered Databases:</b>\n\n")
	for _, db := range dbs {
		sb.WriteString(fmt.Sprintf("<b>%s</b> (%s)\n", db.Name, db.Type))
		sb.WriteString(fmt.Sprintf("Status: <code>%s</code>\n", db.Status))
		sb.WriteString(fmt.Sprintf("UUID: <code>%s</code>\n", db.UUID))
		sb.WriteString("--------------------\n")
	}

	_, err = msg.ReplyText(c, sb.String(), sendOpts(msg, nil))
	return err
}
