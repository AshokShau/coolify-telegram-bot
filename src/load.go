package src

import (
	"fmt"
	"time"

	"coolifymanager/src/scheduler"

	td "github.com/AshokShau/gotdbot"
)

var (
	startTime = time.Now()
)

func registerDevCommand(c *td.Client, cmd string, handler func(*td.Client, *td.Message) error) {
	ch := c.OnCommand(cmd, handler)
	ch.Filter = IsDevMsg
}

func InitFunc(c *td.Client) error {
	if err := scheduler.Start(); err != nil {
		return fmt.Errorf("scheduler start error: %s", err.Error())
	}

	c.OnCommand("start", startHandler)
	c.OnCommand("ping", pingHandler)

	registerDevCommand(c, "jobs", jobsHandler)
	registerDevCommand(c, "job", scheduleHandler)
	registerDevCommand(c, "schedule", scheduleHandler)
	registerDevCommand(c, "unschedule", unscheduleHandler)
	registerDevCommand(c, "rmJob", unscheduleHandler)
	registerDevCommand(c, "apps", appsHandler)
	registerDevCommand(c, "projects", appsHandler)
	registerDevCommand(c, "env", envCmdHandler)
	registerDevCommand(c, "logs", logsCmdHandler)
	registerDevCommand(c, "status", statusCmdHandler)
	registerDevCommand(c, "deployments", deploymentsHandler)
	registerDevCommand(c, "deploy", deployCmdHandler)
	registerDevCommand(c, "servers", serversHandler)
	registerDevCommand(c, "databases", databasesHandler)

	c.OnUpdateNewCallbackQuery(jobsPaginationHandler, DevCallbackPrefix(c, "jobs:"))
	c.OnUpdateNewCallbackQuery(listProjectsHandler, DevCallbackPrefix(c, "list_projects"))
	c.OnUpdateNewCallbackQuery(projectSelectHandler, DevCallbackPrefix(c, "proj:"))
	c.OnUpdateNewCallbackQuery(projectMenuHandler, DevCallbackPrefix(c, "project_menu:"))
	c.OnUpdateNewCallbackQuery(envCallbackHandler, DevCallbackPrefix(c, "env:"))
	c.OnUpdateNewCallbackQuery(scheduleMenuHandler, DevCallbackPrefix(c, "sch_m:"))
	c.OnUpdateNewCallbackQuery(scheduleActionHandler, DevCallbackPrefix(c, "sch_a:"))
	c.OnUpdateNewCallbackQuery(scheduleCreateHandler, DevCallbackPrefix(c, "sch_c:"))
	c.OnUpdateNewCallbackQuery(restartHandler, DevCallbackPrefix(c, "restart:"))
	c.OnUpdateNewCallbackQuery(deployHandler, DevCallbackPrefix(c, "deploy:"))
	c.OnUpdateNewCallbackQuery(logsHandler, DevCallbackPrefix(c, "logs:"))
	c.OnUpdateNewCallbackQuery(statusHandler, DevCallbackPrefix(c, "status:"))
	c.OnUpdateNewCallbackQuery(stopHandler, DevCallbackPrefix(c, "stop:"))
	c.OnUpdateNewCallbackQuery(deleteHandler, DevCallbackPrefix(c, "delete:"))

	return nil
}
