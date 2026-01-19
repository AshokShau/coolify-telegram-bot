package src

import (
	"time"

	"coolifymanager/src/scheduler"

	"github.com/amarnathcjd/gogram/telegram"
)

var (
	startTime = time.Now()
)

func InitFunc(c *telegram.Client) {
	scheduler.Start()
	_, _ = c.UpdatesGetState()

	// Commands
	c.On("command:start", startHandler)
	c.On("command:ping", pingHandler)
	c.On("command:schedule", scheduleHandler)

	//	Callbacks
	c.On("callback:list_projects", listProjectsHandler)
	c.On("callback:list_projects:", listProjectsHandler)
	c.On("callback:project_menu:", projectMenuHandler)
	c.On("callback:restart:", restartHandler)
	c.On("callback:deploy:", deployHandler)
	c.On("callback:logs:", logsHandler)
	c.On("callback:status:", statusHandler)
	c.On("callback:stop:", stopHandler)
	c.On("callback:delete:", deleteHandler)
	c.Logger.Info("Handlers loaded successfully.")
}
