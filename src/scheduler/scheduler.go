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
	s, err = gocron.NewScheduler()
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
		if d, ok := parseDurationSchedule(task.Schedule); ok {
			jobDefinition = gocron.DurationJob(d)
		} else {
			cronExpr := parseSchedule(task.Schedule)
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

	//if task.OneTime {}

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

func parseDurationSchedule(schedule string) (time.Duration, bool) {
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

func parseSchedule(schedule string) string {
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
