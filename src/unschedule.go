package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"strings"

	"github.com/AshokShau/gotdbot"
)

func unscheduleHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	msg := ctx.EffectiveMessage

	if !config.IsDev(msg.SenderID()) {
		_, err := msg.ReplyText(c, "🚫 You are not authorized to use this command.", nil)
		return err
	}

	args := strings.Fields(msg.Text())
	if len(args) < 2 {
		_, err := msg.ReplyText(c, "usage: /unschedule <task_id>", nil)
		return err
	}
	taskID := args[1]

	if err := scheduler.RemoveTask(taskID); err != nil {
		_, err = msg.ReplyText(c, fmt.Sprintf("⚠️ Warning: Could not remove task from scheduler (might not be running): %v", err), nil)
	}

	if err := database.DeleteTask(taskID); err != nil {
		_, err = msg.ReplyText(c, fmt.Sprintf("❌ Error deleting task from database: %v", err), nil)
		return err
	}

	_, err := msg.ReplyText(c, fmt.Sprintf("✅ Task <code>%s</code> removed successfully.", taskID), &gotdbot.SendTextMessageOpts{ParseMode: "HTML"})
	return err
}
