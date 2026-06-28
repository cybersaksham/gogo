package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Route struct {
	Queue      string
	RoutingKey string
	Priority   int
	Headers    map[string]string
}

type RouteFunc func(Signature) (Route, bool)

type RouterOptions struct {
	DefaultQueue   string
	StaticRoutes   map[string]Route
	DynamicRoutes  []RouteFunc
	DefaultHeaders map[string]string
}

type Router struct {
	defaultQueue   string
	staticRoutes   map[string]Route
	dynamicRoutes  []RouteFunc
	defaultHeaders map[string]string
}

func NewRouter(options RouterOptions) *Router {
	if options.DefaultQueue == "" {
		options.DefaultQueue = "default"
	}
	staticRoutes := make(map[string]Route, len(options.StaticRoutes))
	for name, route := range options.StaticRoutes {
		staticRoutes[name] = cloneRoute(route)
	}
	return &Router{
		defaultQueue:   options.DefaultQueue,
		staticRoutes:   staticRoutes,
		dynamicRoutes:  append([]RouteFunc(nil), options.DynamicRoutes...),
		defaultHeaders: cloneStringMap(options.DefaultHeaders),
	}
}

func (r *Router) Route(signature Signature, task Task) Route {
	if r == nil {
		r = NewRouter(RouterOptions{})
	}
	route := Route{Queue: r.defaultQueue, Headers: cloneStringMap(r.defaultHeaders)}
	if task.Options.Queue != "" {
		route.Queue = task.Options.Queue
	}
	if task.Options.RoutingKey != "" {
		route.RoutingKey = task.Options.RoutingKey
	}
	if task.Options.Priority != 0 {
		route.Priority = task.Options.Priority
	}
	if static, ok := r.staticRoutes[signature.Name]; ok {
		route = mergeRoute(route, static)
	}
	for _, dynamic := range r.dynamicRoutes {
		if dynamic == nil {
			continue
		}
		if selected, ok := dynamic(signature.Clone()); ok {
			route = mergeRoute(route, selected)
			break
		}
	}
	if signature.Options.Queue != "" {
		route.Queue = signature.Options.Queue
	}
	if signature.Options.Priority != 0 {
		route.Priority = signature.Options.Priority
	}
	route.Headers = mergeStringMaps(route.Headers, signature.Headers)
	if route.Queue == "" {
		route.Queue = "default"
	}
	return route
}

type SendOptions struct {
	Router        *Router
	ID            string
	RootID        string
	ParentID      string
	GroupID       string
	ChordID       string
	Retries       int
	ReplyTo       string
	CorrelationID string
	CreatedAt     time.Time
}

func (a *App) SendTask(ctx context.Context, broker Broker, signature Signature, options SendOptions) (BrokerMessage, error) {
	if broker == nil {
		return BrokerMessage{}, fmt.Errorf("%w: broker is required", ErrWorkerNotConfigured)
	}
	task, ok := a.Task(signature.Name)
	if !ok {
		return BrokerMessage{}, fmt.Errorf("%w: %s", ErrTaskNotRegistered, signature.Name)
	}
	router := options.Router
	if router == nil {
		router = NewRouter(RouterOptions{DefaultQueue: a.options.DefaultQueue})
	}
	route := router.Route(signature, task)
	signature = signature.Clone()
	signature.Options.Queue = route.Queue
	signature.Options.Priority = route.Priority
	signature.Headers = mergeStringMaps(route.Headers, signature.Headers)
	if route.RoutingKey != "" {
		signature.Headers["routing_key"] = route.RoutingKey
	}
	id := options.ID
	if id == "" {
		id = uuid.NewString()
	}
	envelope := NewEnvelope(signature, EnvelopeOptions{
		ID:            id,
		RootID:        options.RootID,
		ParentID:      options.ParentID,
		GroupID:       options.GroupID,
		ChordID:       options.ChordID,
		Retries:       options.Retries,
		ReplyTo:       options.ReplyTo,
		CorrelationID: options.CorrelationID,
		CreatedAt:     options.CreatedAt,
	})
	return broker.Publish(ctx, route.Queue, envelope, BrokerPublishOptions{Priority: route.Priority, RoutingKey: route.RoutingKey, Headers: route.Headers})
}

func cloneRoute(route Route) Route {
	route.Headers = cloneStringMap(route.Headers)
	return route
}

func mergeRoute(base Route, override Route) Route {
	merged := cloneRoute(base)
	if override.Queue != "" {
		merged.Queue = override.Queue
	}
	if override.RoutingKey != "" {
		merged.RoutingKey = override.RoutingKey
	}
	if override.Priority != 0 {
		merged.Priority = override.Priority
	}
	merged.Headers = mergeStringMaps(merged.Headers, override.Headers)
	return merged
}

func mergeStringMaps(base map[string]string, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return map[string]string{}
	}
	merged := cloneStringMap(base)
	if merged == nil {
		merged = map[string]string{}
	}
	for key, value := range override {
		merged[key] = value
	}
	return merged
}
