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
		_, err := replyMsg(c, msg, "Usage: /schedule <name> <schedule_type> [expression/time]\n"+
			"Types: one_time, every_minute, hourly, daily, weekly, monthly, yearly, cron\n"+
			"For one_time, use RFC3339 format (e.g., 2023-10-27T10:00:00Z)", nil)
		return err
	}

	name := args[1]
	schType := strings.ToLower(args[2])

	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_, err = replyMsg(c, msg, fmt.Sprintf("Error fetching projects: %v", err), nil)
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
		_, err = replyMsg(c, msg, fmt.Sprintf("Project not found with name: %s", name), nil)
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
			_, err = replyMsg(c, msg, "Please provide a time for one-time schedule.", nil)
			return err
		}
		timeStr := args[3]
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			_, err = replyMsg(c, msg, "Invalid time format. Use RFC3339 (e.g., 2023-10-27T10:00:00Z)", nil)
			return err
		}

		if t.Before(time.Now()) {
			_, err = replyMsg(c, msg, "Time must be in the future.", nil)
			return err
		}

		task.OneTime = true
		task.NextRun = t
		task.Schedule = "one_time"

	case "cron":
		if len(args) < 4 {
			_, err = replyMsg(c, msg, "Please provide a cron expression.", nil)
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
				_, err = replyMsg(c, msg, "Invalid time format. Use HH:MM (e.g., 06:00)", nil)
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
					_, err = replyMsg(c, msg, "Invalid time format in schedule. Use HH:MM (e.g., every_1d_at_06:00)", nil)
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
				} else if strings.HasSuffix(base, "d") {
					if _, err := strconv.Atoi(strings.TrimSuffix(base, "d")); err == nil {
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

		if strings.HasSuffix(schType, "d") {
			if _, err := strconv.Atoi(strings.TrimSuffix(schType, "d")); err == nil {
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

		_, err = replyMsg(c, msg, fmt.Sprintf("Unknown schedule type: %s", schType), nil)
		return err
	}

	if err := database.AddTask(task); err != nil {
		_, err = replyMsg(c, msg, fmt.Sprintf("Error saving task: %v", err), nil)
		return err
	}

	if err := scheduler.ScheduleTask(task); err != nil {
		_ = database.DeleteTask(task.ID.Hex())
		_, err = replyMsg(c, msg, fmt.Sprintf("Error scheduling task: %v", err), nil)
		return err
	}

	_, err = replyMsg(c, msg, fmt.Sprintf("Task scheduled successfully!\nID: %s", task.ID.Hex()), nil)
	return err
}
