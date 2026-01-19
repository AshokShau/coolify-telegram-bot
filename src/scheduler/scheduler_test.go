package scheduler

import (
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRemoveTask(t *testing.T) {
	var err error
	s, err = gocron.NewScheduler()
	if err != nil {
		t.Fatal(err)
	}
	s.Start()
	defer func() {
		_ = s.Shutdown()
	}()

	taskID := bson.NewObjectID()
	job, err := s.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(time.Now().Add(1*time.Hour))),
		gocron.NewTask(func() {}),
		gocron.WithTags(taskID.Hex()),
	)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, j := range s.Jobs() {
		if j.ID() == job.ID() {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Job not found after adding")
	}

	if err := RemoveTask(taskID.Hex()); err != nil {
		t.Fatalf("RemoveTask failed: %v", err)
	}
	found = false
	for _, j := range s.Jobs() {
		if j.ID() == job.ID() {
			found = true
			break
		}
	}
	if found {
		t.Fatal("Job still exists after RemoveTask")
	}
}
