package src

import (
	"fmt"
	"strings"
	"time"

	"coolifymanager/src/scheduler"

	"github.com/AshokShau/gotdbot"
	"github.com/AshokShau/gotdbot/ext"
	"github.com/AshokShau/gotdbot/ext/handlers"
	"github.com/AshokShau/gotdbot/ext/handlers/filters"
)

var (
	startTime = time.Now()
)

func CbPrefix(prefix string) filters.CallbackQuery {
	return func(cq *gotdbot.UpdateNewCallbackQuery) bool {

		if cq.CallbackData() == nil {
			return false
		}

		stringData := string(cq.CallbackData())
		return strings.HasPrefix(stringData, prefix)
	}
}

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
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("jobs:"), jobsPaginationHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("list_projects"), listProjectsHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("project_menu:"), projectMenuHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("sch_m:"), scheduleMenuHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("sch_a:"), scheduleActionHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("sch_c:"), scheduleCreateHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("restart:"), restartHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("deploy:"), deployHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("logs:"), logsHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("status:"), statusHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("stop:"), stopHandler))
	d.AddHandler(handlers.NewCallbackQuery(CbPrefix("delete:"), deleteHandler))
	return nil
}
