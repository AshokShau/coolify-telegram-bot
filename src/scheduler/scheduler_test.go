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

func TestParseDurationSchedule(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		ok       bool
	}{
		{"every_48h", 48 * time.Hour, true},
		{"every_6h", 6 * time.Hour, true},
		{"every_30m", 30 * time.Minute, true},
		{"every_2d", 48 * time.Hour, true},
		{"every_10s", 10 * time.Second, true},
		{"every_minute", 0, false},
		{"hourly", 0, false},
		{"invalid", 0, false},
		{"every_invalid", 0, false},
	}

	for _, tt := range tests {
		d, ok := ParseDurationSchedule(tt.input)
		if ok != tt.ok {
			t.Errorf("ParseDurationSchedule(%q) ok = %v, want %v", tt.input, ok, tt.ok)
		}
		if ok && d != tt.expected {
			t.Errorf("ParseDurationSchedule(%q) duration = %v, want %v", tt.input, d, tt.expected)
		}
	}
}
