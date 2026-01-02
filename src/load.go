package src

import (
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

var (
	startTime = time.Now()
)

func InitFunc(c *telegram.Client) {
	_, _ = c.UpdatesGetState()

	// Commands
	c.On("command:start", startHandler)
	c.On("command:ping", pingHandler)

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
