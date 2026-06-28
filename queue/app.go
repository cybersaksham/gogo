package queue

import (
	"fmt"
	"sort"
	"sync"
)

// AppOptions configures the queue app defaults.
type AppOptions struct {
	DefaultQueue      string
	DefaultSerializer string
	DefaultMaxRetries int
	DefaultAckPolicy  AckPolicy
}

// App owns task registration and lookup.
type App struct {
	mu      sync.RWMutex
	options AppOptions
	tasks   map[string]Task
}

func NewApp(options AppOptions) *App {
	if options.DefaultQueue == "" {
		options.DefaultQueue = "default"
	}
	if options.DefaultSerializer == "" {
		options.DefaultSerializer = "json"
	}
	if options.DefaultAckPolicy == "" {
		options.DefaultAckPolicy = AckEarly
	}
	return &App{
		options: options,
		tasks:   map[string]Task{},
	}
}

func (a *App) RegisterTask(name string, fn TaskFunc, options TaskOptions) (Task, error) {
	if err := validateTask(name, fn); err != nil {
		return Task{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.tasks[name]; exists {
		return Task{}, fmt.Errorf("%w: %s", ErrDuplicateTask, name)
	}
	task := Task{Name: name, Func: fn, Options: a.withDefaults(options)}
	a.tasks[name] = task
	return task, nil
}

func (a *App) DiscoverTasks(providers ...TaskProvider) error {
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		for _, definition := range provider.QueueTasks() {
			if _, err := a.RegisterTask(definition.Name, definition.Func, definition.Options); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) Task(name string) (Task, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	task, ok := a.tasks[name]
	return task, ok
}

func (a *App) Tasks() []Task {
	a.mu.RLock()
	defer a.mu.RUnlock()
	names := make([]string, 0, len(a.tasks))
	for name := range a.tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	tasks := make([]Task, len(names))
	for i, name := range names {
		tasks[i] = a.tasks[name]
	}
	return tasks
}

func (a *App) withDefaults(options TaskOptions) TaskOptions {
	if options.Serializer == "" {
		options.Serializer = a.options.DefaultSerializer
	}
	if options.Queue == "" {
		options.Queue = a.options.DefaultQueue
	}
	if options.MaxRetries == 0 {
		options.MaxRetries = a.options.DefaultMaxRetries
	}
	if options.AckPolicy == "" {
		options.AckPolicy = a.options.DefaultAckPolicy
	}
	return options
}
