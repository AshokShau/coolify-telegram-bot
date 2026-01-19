package src

import (
	"fmt"
	"time"

	"coolifymanager/src/scheduler"

	"github.com/amarnathcjd/gogram/telegram"
)

var (
	startTime = time.Now()
)

func InitFunc(c *telegram.Client) error {
	if err := scheduler.Start(); err != nil {
		return fmt.Errorf("scheduler start error: %s", err.Error())
	}

	_, _ = c.UpdatesGetState()

	// Commands
	c.On("command:start", startHandler)
	c.On("command:ping", pingHandler)
	c.On("command:jobs", jobsHandler)
	c.On("command:job", scheduleHandler)
	c.On("command:schedule", scheduleHandler)
	c.On("command:unschedule", unscheduleHandler)
	c.On("command:rmJob", unscheduleHandler)

	//	Callbacks
	c.On("callback:jobs:", jobsPaginationHandler)
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
	return nil
}
