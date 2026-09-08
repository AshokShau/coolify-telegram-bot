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
		_, err := replyMsg(c, msg, "Usage: /unschedule <task_id>", nil)
		return err
	}
	taskID := args[1]

	if err := scheduler.RemoveTask(taskID); err != nil {
		_, _ = replyMsg(c, msg, fmt.Sprintf("Warning: Could not remove task from scheduler: %v", err), nil)
	}

	if err := database.DeleteTask(taskID); err != nil {
		_, err = replyMsg(c, msg, fmt.Sprintf("Error deleting task from database: %v", err), nil)
		return err
	}

	_, err := replyMsg(c, msg, fmt.Sprintf("Task <code>%s</code> removed successfully.", taskID), nil)
	return err
}
