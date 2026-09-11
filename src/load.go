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

	c.OnUpdateNewCallbackQuery(startMenuCallbackHandler, DevCallbackPrefix(c, "start_menu"))
	c.OnUpdateNewCallbackQuery(jobsPaginationHandler, DevCallbackPrefix(c, "jobs:"))
	c.OnUpdateNewCallbackQuery(unscheduleCallbackHandler, DevCallbackPrefix(c, "unschedule:"))
	c.OnUpdateNewCallbackQuery(deploymentsCallbackHandler, DevCallbackPrefix(c, "deployments:"))
	c.OnUpdateNewCallbackQuery(serversCallbackHandler, DevCallbackPrefix(c, "servers:"))
	c.OnUpdateNewCallbackQuery(databasesCallbackHandler, DevCallbackPrefix(c, "databases:"))
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
