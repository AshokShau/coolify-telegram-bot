package scheduler

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
)

var s gocron.Scheduler

func Start() error {
	var err error
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Printf("Error loading time zone: %v, falling back to Local", err)
		loc = time.Local
	}

	s, err = gocron.NewScheduler(
		gocron.WithLocation(loc),
	)
	if err != nil {
		return fmt.Errorf("error initializing scheduler: %w", err)
	}

	tasks, err := database.GetTasks()
	if err != nil {
		log.Printf("Error loading tasks: %v", err)
	}

	for _, task := range tasks {
		if err := ScheduleTask(task); err != nil {
			log.Printf("Error scheduling task %v: %v", task.ID, err)
		}
	}

	s.Start()
	log.Println("Scheduler started")
	return nil
}

func Shutdown() error {
	if s != nil {
		return s.Shutdown()
	}
	return nil
}

func ScheduleTask(task database.ScheduledTask) error {
	var jobDefinition gocron.JobDefinition

	if task.OneTime {
		if task.NextRun.Before(time.Now()) {
			log.Printf("Task %v (OneTime) is in the past, skipping and removing.", task.ID)
			_ = database.RemoveOneTimeTask(task.ID)
			return nil
		}
		jobDefinition = gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(task.NextRun),
		)
	} else {
		isDaySchedule := false
		if strings.HasPrefix(task.Schedule, "every_") && strings.HasSuffix(task.Schedule, "d") {
			s := strings.TrimSuffix(strings.TrimPrefix(task.Schedule, "every_"), "d")
			if _, err := strconv.Atoi(s); err == nil {
				isDaySchedule = true
			}
		}

		if isDaySchedule || strings.Contains(task.Schedule, "_at_") {
			cronExpr := parseSchedule(task.Schedule, task.ID.Timestamp())
			jobDefinition = gocron.CronJob(cronExpr, false)
		} else if d, ok := ParseDurationSchedule(task.Schedule); ok {
			jobDefinition = gocron.DurationJob(d)
		} else {
			cronExpr := parseSchedule(task.Schedule, task.ID.Timestamp())
			jobDefinition = gocron.CronJob(
				cronExpr,
				false,
			)
		}
	}

	job, err := s.NewJob(
		jobDefinition,
		gocron.NewTask(executeTask, task),
		gocron.WithTags(task.ID.Hex()),
	)
	if err != nil {
		return err
	}

	log.Printf("Scheduled job %s for task %v", job.ID(), task.ID)
	return nil
}

func RemoveTask(id string) error {
	for _, j := range s.Jobs() {
		for _, tag := range j.Tags() {
			if tag == id {
				return s.RemoveJob(j.ID())
			}
		}
	}
	return nil
}

func ParseDurationSchedule(schedule string) (time.Duration, bool) {
	if !strings.HasPrefix(schedule, "every_") {
		return 0, false
	}
	s := strings.TrimPrefix(schedule, "every_")
	if strings.HasSuffix(s, "d") {
		val, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, false
		}
		return time.Duration(val) * 24 * time.Hour, true
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, false
	}
	return d, true
}

func parseSchedule(schedule string, t time.Time) string {
	if strings.Contains(schedule, "_at_") {
		parts := strings.Split(schedule, "_at_")
		if len(parts) == 2 {
			base := parts[0]
			timeStr := parts[1]
			timeParts := strings.Split(timeStr, ":")
			if len(timeParts) == 2 {
				hour, errH := strconv.Atoi(timeParts[0])
				minute, errM := strconv.Atoi(timeParts[1])
				if errH == nil && errM == nil {
					switch base {
					case "daily":
						return fmt.Sprintf("%d %d * * *", minute, hour)
					default:
						if strings.HasPrefix(base, "every_") && strings.HasSuffix(base, "d") {
							s := strings.TrimPrefix(base, "every_")
							val, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
							if err == nil {
								if val == 1 {
									return fmt.Sprintf("%d %d * * *", minute, hour)
								}
								return fmt.Sprintf("%d %d */%d * *", minute, hour, val)
							}
						}
					}
				}
			}
		}
	}

	switch schedule {
	case "every_minute":
		return "* * * * *"
	case "hourly":
		return "0 * * * *"
	case "daily":
		return "0 0 * * *"
	case "weekly":
		return "0 0 * * 0"
	case "monthly":
		return "0 0 1 * *"
	case "yearly":
		return "0 0 1 1 *"
	default:
		if strings.HasPrefix(schedule, "every_") && strings.HasSuffix(schedule, "d") {
			s := strings.TrimPrefix(schedule, "every_")
			val, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
			if err == nil {
				minute := t.Minute()
				hour := t.Hour()
				if val == 1 {
					return fmt.Sprintf("%d %d * * *", minute, hour)
				}
				return fmt.Sprintf("%d %d */%d * *", minute, hour, val)
			}
		}
		return schedule
	}
}

func executeTask(task database.ScheduledTask) {
	log.Printf("Executing task '%s' (%v) for project %s", task.Name, task.ID, task.ProjectUUID)

	if task.Type == "restart" {
		_, err := config.Coolify.RestartApplicationByUUID(task.ProjectUUID)
		if err != nil {
			log.Printf("Error restarting project %s (Task: %s): %v", task.ProjectUUID, task.Name, err)
		} else {
			log.Printf("Successfully restarted project %s (Task: %s)", task.ProjectUUID, task.Name)
		}
	} else {
		log.Printf("Unknown task type: %s (Task: %s)", task.Type, task.Name)
	}

	if task.OneTime {
		if err := database.RemoveOneTimeTask(task.ID); err != nil {
			log.Printf("Error removing executed one-time task %v: %v", task.ID, err)
		}
	}
}
