package src

import (
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"strings"

	td "github.com/AshokShau/gotdbot"
)

func unscheduleHandler(c *td.Client, msg *td.Message) error {
	args := strings.Fields(msg.Text())
	if len(args) < 2 {
		_, err := msg.ReplyText(c, "Usage: /unschedule <task_id>", sendOpts(msg, nil))
		return err
	}
	taskID := args[1]

	if err := scheduler.RemoveTask(taskID); err != nil {
		_, _ = msg.ReplyText(c, fmt.Sprintf("Warning: Could not remove task from scheduler: %v", err), sendOpts(msg, nil))
	}

	if err := database.DeleteTask(taskID); err != nil {
		_, err = msg.ReplyText(c, fmt.Sprintf("Error deleting task from database: %v", err), sendOpts(msg, nil))
		return err
	}

	_, err := msg.ReplyText(c, fmt.Sprintf("Task <code>%s</code> removed successfully.", taskID), sendOpts(msg, nil))
	return err
}
