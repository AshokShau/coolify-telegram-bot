package src

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
)

func errorHandler(bot *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
	var msg string
	if ctx.Update != nil {
		if updateBytes, err := json.MarshalIndent(ctx.Update, "", "  "); err == nil {
			msg = html.EscapeString(string(updateBytes))
		} else {
			msg = "failed to marshal update"
		}
	} else {

		msg = "no update"
	}

	message := fmt.Sprintf("<blockquote expandable>New Error:\n%s\n\n%s</blockquote>", err.Error(), msg)
	if _, err = bot.SendMessage(5938660179, message, &gotgbot.SendMessageOpts{ParseMode: "HTML", DisableNotification: true}); err != nil {
		log.Printf("failed to send error message to logger: %s", err)
		return ext.DispatcherActionNoop
	}

	return ext.DispatcherActionNoop
}

var (
	startTime  = time.Now()
	Dispatcher = newDispatcher()
)

func newDispatcher() *ext.Dispatcher {
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{Error: errorHandler, MaxRoutines: -1})
	dispatcher.AddHandler(handlers.NewCommand("start", startHandler))
	dispatcher.AddHandler(handlers.NewCommand("ping", pingCommandHandler))

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("list_projects"), listProjectsHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("project_menu:"), projectMenuHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("restart:"), restartHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("deploy:"), deployHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("logs:"), logsHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("status:"), statusHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("stop:"), stopHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("delete:"), deleteHandler))
	return dispatcher
}
