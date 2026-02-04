package src

import (
	"fmt"
	"time"

	"coolifymanager/src/scheduler"

	"github.com/AshokShau/gotdbot/ext"
	"github.com/AshokShau/gotdbot/ext/handlers"
	"github.com/AshokShau/gotdbot/ext/handlers/filters/callbackquery"
)

var (
	startTime = time.Now()
)

func InitFunc(d *ext.Dispatcher) error {
	if err := scheduler.Start(); err != nil {
		return fmt.Errorf("scheduler start error: %s", err.Error())
	}

	// Commands
	d.AddHandler(handlers.NewCommand("start", startHandler))
	d.AddHandler(handlers.NewCommand("ping", pingHandler))
	d.AddHandler(handlers.NewCommand("jobs", jobsHandler))
	d.AddHandler(handlers.NewCommand("job", scheduleHandler))
	d.AddHandler(handlers.NewCommand("schedule", scheduleHandler))
	d.AddHandler(handlers.NewCommand("unschedule", unscheduleHandler))
	d.AddHandler(handlers.NewCommand("rmJob", unscheduleHandler))

	//	Callbacks
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("jobs:"), jobsPaginationHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("list_projects"), listProjectsHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("project_menu:"), projectMenuHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("sch_m:"), scheduleMenuHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("sch_a:"), scheduleActionHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("sch_c:"), scheduleCreateHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("restart:"), restartHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("deploy:"), deployHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("logs:"), logsHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("status:"), statusHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("stop:"), stopHandler))
	d.AddHandler(handlers.NewUpdateNewCallbackQuery(callbackquery.Prefix("delete:"), deleteHandler))
	return nil
}
