package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func unscheduleHandler(m *telegram.NewMessage) error {
	if !config.IsDev(m.Sender.ID) {
		_, err := m.Reply("🚫 You are not authorized to use this command.")
		return err
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		_, err := m.Reply("usage: /unschedule <task_id>")
		return err
	}

	taskID := args[1]

	if err := scheduler.RemoveTask(taskID); err != nil {
		_, err = m.Reply(fmt.Sprintf("⚠️ Warning: Could not remove task from scheduler (might not be running): %v", err))
	}
	
	if err := database.DeleteTask(taskID); err != nil {
		_, err = m.Reply(fmt.Sprintf("❌ Error deleting task from database: %v", err))
		return err
	}

	_, err := m.Reply(fmt.Sprintf("✅ Task <code>%s</code> removed successfully.", taskID))
	return err
}
