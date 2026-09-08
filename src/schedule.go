package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"strconv"
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func scheduleHandler(c *td.Client, msg *td.Message) error {
	args := strings.Fields(msg.Text())
	if len(args) < 3 {
		_, err := msg.ReplyText(c, "Usage: /schedule <name> <schedule_type> [expression/time]\n"+
			"Types: one_time, every_minute, hourly, daily, weekly, monthly, yearly, cron\n"+
			"For one_time, use RFC3339 format (e.g., 2023-10-27T10:00:00Z)", sendOpts(msg, nil))
		return err
	}

	name := args[1]
	schType := strings.ToLower(args[2])

	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_, err = msg.ReplyText(c, fmt.Sprintf("Error fetching projects: %v", err), sendOpts(msg, nil))
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
		_, err = msg.ReplyText(c, fmt.Sprintf("Project not found with name: %s", name), sendOpts(msg, nil))
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
			_, err = msg.ReplyText(c, "Please provide a time for one-time schedule.", sendOpts(msg, nil))
			return err
		}
		timeStr := args[3]
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			_, err = msg.ReplyText(c, "Invalid time format. Use RFC3339 (e.g., 2023-10-27T10:00:00Z)", sendOpts(msg, nil))
			return err
		}

		if t.Before(time.Now()) {
			_, err = msg.ReplyText(c, "Time must be in the future.", sendOpts(msg, nil))
			return err
		}

		task.OneTime = true
		task.NextRun = t
		task.Schedule = "one_time"

	case "cron":
		if len(args) < 4 {
			_, err = msg.ReplyText(c, "Please provide a cron expression.", sendOpts(msg, nil))
			return err
		}

		cronExpr := strings.Join(args[3:], " ")
		task.Schedule = cronExpr

	case "every_minute", "hourly", "weekly", "monthly", "yearly":
		task.Schedule = schType

	case "daily":
		if len(args) >= 4 {
			timeStr := args[3]
			if _, err := time.Parse("15:04", timeStr); err == nil {
				task.Schedule = "daily_at_" + timeStr
			} else {
				_, err = msg.ReplyText(c, "Invalid time format. Use HH:MM (e.g., 06:00)", sendOpts(msg, nil))
				return err
			}
		} else {
			task.Schedule = "daily"
		}

	default:
		if strings.Contains(schType, "_at_") {
			parts := strings.Split(schType, "_at_")
			if len(parts) == 2 {
				base := parts[0]
				timeStr := parts[1]
				if _, err := time.Parse("15:04", timeStr); err != nil {
					_, err = msg.ReplyText(c, "Invalid time format in schedule. Use HH:MM (e.g., every_1d_at_06:00)", sendOpts(msg, nil))
					return err
				}

				if base == "daily" {
					task.Schedule = schType
					break
				} else if strings.HasPrefix(base, "every_") && strings.HasSuffix(base, "d") {
					if _, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(base, "every_"), "d")); err == nil {
						task.Schedule = schType
						break
					}
				} else if before, ok := strings.CutSuffix(base, "d"); ok {
					if _, err := strconv.Atoi(before); err == nil {
						task.Schedule = "every_" + base + "_at_" + timeStr
						break
					}
				}
			}
		}

		if _, ok := scheduler.ParseDurationSchedule(schType); ok {
			task.Schedule = schType
			break
		}

		if before, ok := strings.CutSuffix(schType, "d"); ok {
			if _, err := strconv.Atoi(before); err == nil {
				if len(args) >= 4 {
					timeStr := args[3]
					if _, err := time.Parse("15:04", timeStr); err == nil {
						task.Schedule = "every_" + schType + "_at_" + timeStr
						break
					}
				}
				task.Schedule = "every_" + schType
				break
			}
		}

		if _, err := time.ParseDuration(schType); err == nil {
			task.Schedule = "every_" + schType
			break
		}

		_, err = msg.ReplyText(c, fmt.Sprintf("Unknown schedule type: %s", schType), sendOpts(msg, nil))
		return err
	}

	if err := database.AddTask(task); err != nil {
		_, err = msg.ReplyText(c, fmt.Sprintf("Error saving task: %v", err), sendOpts(msg, nil))
		return err
	}

	if err := scheduler.ScheduleTask(task); err != nil {
		_ = database.DeleteTask(task.ID.Hex())
		_, err = msg.ReplyText(c, fmt.Sprintf("Error scheduling task: %v", err), sendOpts(msg, nil))
		return err
	}

	_, err = msg.ReplyText(c, fmt.Sprintf("Task scheduled successfully!\nID: %s", task.ID.Hex()), sendOpts(msg, nil))
	return err
}
