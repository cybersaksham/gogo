package queue

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestQueueTaskRegistryRegistrationDefaultsDuplicatesAndDiscovery(t *testing.T) {
	app := NewApp(AppOptions{
		DefaultQueue:      "default",
		DefaultSerializer: "json",
		DefaultMaxRetries: 3,
		DefaultAckPolicy:  AckLate,
	})

	task, err := app.RegisterTask("blog.publish", testTaskFunc, TaskOptions{
		Serializer:        "gob",
		Queue:             "emails",
		RoutingKey:        "blog.publish",
		Priority:          7,
		MaxRetries:        5,
		DefaultRetryDelay: 30 * time.Second,
		RetryBackoff:      true,
		RetryJitter:       true,
		SoftTimeout:       time.Second,
		HardTimeout:       2 * time.Second,
		RateLimit:         RateLimit{Limit: 10, Period: time.Minute},
		AckPolicy:         AckManual,
		IgnoreResult:      true,
		TrackStarted:      true,
	})
	if err != nil {
		t.Fatalf("RegisterTask() error = %v", err)
	}
	if task.Name != "blog.publish" || task.Options.Serializer != "gob" || task.Options.Queue != "emails" || task.Options.RoutingKey != "blog.publish" {
		t.Fatalf("task routing/options = %#v", task)
	}
	if task.Options.Priority != 7 || task.Options.MaxRetries != 5 || task.Options.DefaultRetryDelay != 30*time.Second {
		t.Fatalf("task retry options = %#v", task.Options)
	}
	if !task.Options.RetryBackoff || !task.Options.RetryJitter || task.Options.SoftTimeout != time.Second || task.Options.HardTimeout != 2*time.Second {
		t.Fatalf("task timeout/retry flags = %#v", task.Options)
	}
	if task.Options.RateLimit.Limit != 10 || task.Options.RateLimit.Period != time.Minute || task.Options.AckPolicy != AckManual || !task.Options.IgnoreResult || !task.Options.TrackStarted {
		t.Fatalf("task execution options = %#v", task.Options)
	}

	defaulted, err := app.RegisterTask("blog.defaulted", testTaskFunc, TaskOptions{})
	if err != nil {
		t.Fatalf("RegisterTask(defaulted) error = %v", err)
	}
	if defaulted.Options.Serializer != "json" || defaulted.Options.Queue != "default" || defaulted.Options.MaxRetries != 3 || defaulted.Options.AckPolicy != AckLate {
		t.Fatalf("defaulted options = %#v", defaulted.Options)
	}
	if _, err := app.RegisterTask("blog.publish", testTaskFunc, TaskOptions{}); !errors.Is(err, ErrDuplicateTask) {
		t.Fatalf("duplicate error = %v, want ErrDuplicateTask", err)
	}

	err = app.DiscoverTasks(staticTaskProvider{tasks: []TaskDefinition{
		{Name: "blog.discovered", Func: testTaskFunc, Options: TaskOptions{Queue: "discovered"}},
	}})
	if err != nil {
		t.Fatalf("DiscoverTasks() error = %v", err)
	}
	if _, ok := app.Task("blog.discovered"); !ok {
		t.Fatal("discovered task missing")
	}
	if got := taskNames(app.Tasks()); !reflect.DeepEqual(got, []string{"blog.defaulted", "blog.discovered", "blog.publish"}) {
		t.Fatalf("Tasks() names = %#v", got)
	}
}

func TestQueueTaskRegistryRejectsInvalidTasks(t *testing.T) {
	app := NewApp(AppOptions{})
	if _, err := app.RegisterTask("", testTaskFunc, TaskOptions{}); !errors.Is(err, ErrInvalidTask) {
		t.Fatalf("empty name error = %v, want ErrInvalidTask", err)
	}
	if _, err := app.RegisterTask("missing.func", nil, TaskOptions{}); !errors.Is(err, ErrInvalidTask) {
		t.Fatalf("nil func error = %v, want ErrInvalidTask", err)
	}
}

func testTaskFunc(context.Context, ...any) (any, error) {
	return "ok", nil
}

type staticTaskProvider struct {
	tasks []TaskDefinition
}

func (p staticTaskProvider) QueueTasks() []TaskDefinition {
	return append([]TaskDefinition(nil), p.tasks...)
}

func taskNames(tasks []Task) []string {
	names := make([]string, len(tasks))
	for i, task := range tasks {
		names[i] = task.Name
	}
	return names
}
