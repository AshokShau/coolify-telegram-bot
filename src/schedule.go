package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func scheduleHandler(m *telegram.NewMessage) error {
	if !config.IsDev(m.Sender.ID) {
		_, err := m.Reply("🚫 You are not authorized to use this command.")
		return err
	}

	args := strings.Fields(m.Text())
	if len(args) < 3 {
		_, err := m.Reply("usage: /schedule <name> <schedule_type> [expression/time]\n" +
			"Types: one_time, every_minute, hourly, daily, weekly, monthly, yearly, cron\n" +
			"For one_time, use RFC3339 format (e.g., 2023-10-27T10:00:00Z)")
		return err
	}

	name := args[1]
	schType := strings.ToLower(args[2])

	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_, err = m.Reply(fmt.Sprintf("❌ Error fetching projects: %v", err))
		return err
	}

	var uuid string
	for _, app := range apps {
		if strings.EqualFold(app.Name, name) {
			uuid = app.UUID
			break
		}
	}

	if uuid == "" {
		_, err = m.Reply(fmt.Sprintf("❌ Project not found with name: %s", name))
		return err
	}

	task := database.ScheduledTask{
		ID:          bson.NewObjectID(),
		Name:        name,
		ProjectUUID: uuid,
		Type:        "restart",
	}

	switch schType {
	case "one_time":
		if len(args) < 4 {
			_, err = m.Reply("❌ Please provide a time for one-time schedule.")
			return err
		}
		timeStr := args[3]
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			_, err = m.Reply("❌ Invalid time format. Use RFC3339 (e.g., 2023-10-27T10:00:00Z)")
			return err
		}
		if t.Before(time.Now()) {
			_, err = m.Reply("❌ Time must be in the future.")
			return err
		}
		task.OneTime = true
		task.NextRun = t
		task.Schedule = "one_time"

	case "cron":
		if len(args) < 4 {
			_, err = m.Reply("❌ Please provide a cron expression.")
			return err
		}

		cronExpr := strings.Join(args[3:], " ")
		task.Schedule = cronExpr

	case "every_minute", "hourly", "daily", "weekly", "monthly", "yearly":
		task.Schedule = schType

	default:
		if _, ok := scheduler.ParseDurationSchedule(schType); ok {
			task.Schedule = schType
			break
		}
		if _, err := time.ParseDuration(schType); err == nil {
			task.Schedule = "every_" + schType
			break
		}

		_, err = m.Reply(fmt.Sprintf("❌ Unknown schedule type: %s", schType))
		return err
	}

	if err := database.AddTask(task); err != nil {
		_, err = m.Reply(fmt.Sprintf("❌ Error saving task: %v", err))
		return err
	}

	if err := scheduler.ScheduleTask(task); err != nil {
		_ = database.DeleteTask(task.ID.Hex())
		_, err = m.Reply(fmt.Sprintf("❌ Error scheduling task: %v", err))
		return err
	}

	_, err = m.Reply(fmt.Sprintf("✅ Task scheduled successfully!\nID: %s", task.ID.Hex()))
	return err
}
