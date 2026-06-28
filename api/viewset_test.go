package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestAPIViewLifecycleSupportsFunctionAndStructViews(t *testing.T) {
	order := []string{}
	raw := httptest.NewRequest(http.MethodPost, "/posts/", strings.NewReader(`{"title":"Gogo"}`))
	raw.Header.Set("Content-Type", "application/json")

	view := APIView{
		ParseBody: true,
		InitializeRequest: func(_ context.Context, request *Request) (*Request, error) {
			order = append(order, "initialize")
			return request.WithVersion("v1"), nil
		},
		CheckAuthentication: func(context.Context, *Request) error {
			order = append(order, "auth")
			return nil
		},
		CheckPermissions: func(context.Context, *Request) error {
			order = append(order, "permissions")
			return nil
		},
		CheckThrottles: func(context.Context, *Request) error {
			order = append(order, "throttles")
			return nil
		},
		Handler: func(_ context.Context, request *Request) Response {
			order = append(order, "handler")
			return JSON(http.StatusOK, map[string]any{
				"body":    request.ParsedBody(),
				"version": request.Version(),
			})
		},
		FinalizeResponse: func(_ context.Context, _ *Request, response Response) Response {
			order = append(order, "finalize")
			return response
		},
	}

	response := view.AsView()(context.Background(), NewRequest(raw))
	if response.status != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.status)
	}
	if !reflect.DeepEqual(order, []string{"initialize", "auth", "permissions", "throttles", "handler", "finalize"}) {
		t.Fatalf("order = %#v", order)
	}
	body := response.body.(map[string]any)
	if body["version"] != "v1" || body["body"].(map[string]any)["title"] != "Gogo" {
		t.Fatalf("response body = %#v", body)
	}

	functionResponse := FunctionView(func(context.Context, *Request) Response {
		return JSON(http.StatusAccepted, map[string]any{"ok": true})
	})(context.Background(), NewRequest(httptest.NewRequest(http.MethodGet, "/", nil)))
	if functionResponse.status != http.StatusAccepted {
		t.Fatalf("function response status = %d, want 202", functionResponse.status)
	}
}

func TestAPIViewLifecycleFailuresUseExceptionHandler(t *testing.T) {
	tests := []struct {
		name   string
		view   APIView
		raw    *http.Request
		status int
		code   string
	}{
		{
			name: "authentication",
			view: APIView{
				CheckAuthentication: func(context.Context, *Request) error { return ErrAuthenticationFailed },
				Handler:             func(context.Context, *Request) Response { return NoContent() },
			},
			raw:    httptest.NewRequest(http.MethodGet, "/", nil),
			status: http.StatusUnauthorized,
			code:   "authentication_failed",
		},
		{
			name: "permission",
			view: APIView{
				CheckPermissions: func(context.Context, *Request) error { return ErrPermissionDenied },
				Handler:          func(context.Context, *Request) Response { return NoContent() },
			},
			raw:    httptest.NewRequest(http.MethodGet, "/", nil),
			status: http.StatusForbidden,
			code:   "permission_denied",
		},
		{
			name: "throttle",
			view: APIView{
				CheckThrottles: func(context.Context, *Request) error { return ErrThrottled },
				Handler:        func(context.Context, *Request) Response { return NoContent() },
			},
			raw:    httptest.NewRequest(http.MethodGet, "/", nil),
			status: http.StatusTooManyRequests,
			code:   "throttled",
		},
		{
			name: "parse",
			view: APIView{
				ParseBody: true,
				Handler:   func(context.Context, *Request) Response { return NoContent() },
			},
			raw: func() *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{bad"))
				request.Header.Set("Content-Type", "application/json")
				return request
			}(),
			status: http.StatusBadRequest,
			code:   "parse_error",
		},
		{
			name: "panic",
			view: APIView{
				Handler: func(context.Context, *Request) Response { panic("boom") },
			},
			raw:    httptest.NewRequest(http.MethodGet, "/", nil),
			status: http.StatusInternalServerError,
			code:   "internal_error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := test.view.AsView()(context.Background(), NewRequest(test.raw))
			if response.status != test.status {
				t.Fatalf("status = %d, want %d", response.status, test.status)
			}
			body := response.body.(map[string]APIError)
			if body["error"].Code != test.code {
				t.Fatalf("code = %q, want %q", body["error"].Code, test.code)
			}
		})
	}

	custom := APIView{
		CheckPermissions: func(context.Context, *Request) error { return errors.New("hidden") },
		Handler:          func(context.Context, *Request) Response { return NoContent() },
		HandleException: func(_ context.Context, _ *Request, err error) Response {
			return Error(http.StatusTeapot, APIError{Code: "custom", Message: err.Error()})
		},
	}
	response := custom.AsView()(context.Background(), NewRequest(httptest.NewRequest(http.MethodGet, "/", nil)))
	if response.status != http.StatusTeapot {
		t.Fatalf("custom status = %d, want 418", response.status)
	}
}

func TestModelViewSetCRUDAndCustomActions(t *testing.T) {
	store := newMemoryViewSetStore()
	viewset := &ModelViewSet{
		Store: store,
		Serializer: NewSerializer(
			StringField("id", FieldOptions{ReadOnly: true}),
			StringField("title", FieldOptions{Required: true}),
			StringField("status", FieldOptions{}),
		),
	}
	viewset.RegisterAction("publish", ViewSetAction{
		Handler: func(context.Context, *Request) Response {
			return JSON(http.StatusOK, map[string]any{"published": true})
		},
		Detail: true,
	})

	list := viewset.AsView("list")(context.Background(), NewRequest(httptest.NewRequest(http.MethodGet, "/posts/", nil)))
	if list.status != http.StatusOK || len(list.body.([]map[string]any)) != 1 {
		t.Fatalf("list response = %#v", list)
	}

	retrieveRequest := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/1/", nil)).WithPathParam("id", "1")
	retrieve := viewset.AsView("retrieve")(context.Background(), retrieveRequest)
	if retrieve.body.(map[string]any)["title"] != "First" {
		t.Fatalf("retrieve body = %#v", retrieve.body)
	}

	createRequest := NewRequest(httptest.NewRequest(http.MethodPost, "/posts/", nil)).WithParsedBody(map[string]any{"title": "Second", "status": "draft"})
	create := viewset.AsView("create")(context.Background(), createRequest)
	if create.status != http.StatusCreated || create.body.(map[string]any)["id"] != "2" {
		t.Fatalf("create response = %#v", create)
	}

	updateRequest := NewRequest(httptest.NewRequest(http.MethodPut, "/posts/2/", nil)).
		WithPathParam("id", "2").
		WithParsedBody(map[string]any{"title": "Updated", "status": "live"})
	update := viewset.AsView("update")(context.Background(), updateRequest)
	if update.body.(map[string]any)["title"] != "Updated" {
		t.Fatalf("update body = %#v", update.body)
	}

	partialRequest := NewRequest(httptest.NewRequest(http.MethodPatch, "/posts/2/", nil)).
		WithPathParam("id", "2").
		WithParsedBody(map[string]any{"status": "archived"})
	partial := viewset.AsView("partial_update")(context.Background(), partialRequest)
	if partial.body.(map[string]any)["title"] != "Updated" || partial.body.(map[string]any)["status"] != "archived" {
		t.Fatalf("partial body = %#v", partial.body)
	}

	custom := viewset.AsView("publish")(context.Background(), retrieveRequest)
	if custom.status != http.StatusOK || custom.body.(map[string]any)["published"] != true {
		t.Fatalf("custom action = %#v", custom)
	}

	destroy := viewset.AsView("destroy")(context.Background(), retrieveRequest)
	if destroy.status != http.StatusNoContent || store.deleted != "1" {
		t.Fatalf("destroy response = %#v deleted=%q", destroy, store.deleted)
	}

	invalidCreate := viewset.AsView("create")(context.Background(), NewRequest(httptest.NewRequest(http.MethodPost, "/posts/", nil)).WithParsedBody(map[string]any{}))
	if invalidCreate.status != http.StatusBadRequest {
		t.Fatalf("invalid create status = %d, want 400", invalidCreate.status)
	}
}

type memoryViewSetStore struct {
	items   map[string]map[string]any
	nextID  int
	deleted string
}

func newMemoryViewSetStore() *memoryViewSetStore {
	return &memoryViewSetStore{
		items: map[string]map[string]any{
			"1": {"id": "1", "title": "First", "status": "draft"},
		},
		nextID: 2,
	}
}

func (s *memoryViewSetStore) List(context.Context, *Request) ([]map[string]any, error) {
	return []map[string]any{s.items["1"]}, nil
}

func (s *memoryViewSetStore) Retrieve(_ context.Context, _ *Request, lookup string) (map[string]any, error) {
	if item, ok := s.items[lookup]; ok {
		return item, nil
	}
	return nil, ErrNotFound
}

func (s *memoryViewSetStore) Create(_ context.Context, _ *Request, data map[string]any) (map[string]any, error) {
	id := "2"
	data["id"] = id
	s.items[id] = data
	return data, nil
}

func (s *memoryViewSetStore) Update(_ context.Context, _ *Request, lookup string, data map[string]any, partial bool) (map[string]any, error) {
	item := cloneAnyMap(s.items[lookup])
	if !partial {
		item = map[string]any{"id": lookup}
	}
	for key, value := range data {
		item[key] = value
	}
	s.items[lookup] = item
	return item, nil
}

func (s *memoryViewSetStore) Destroy(_ context.Context, _ *Request, lookup string) error {
	s.deleted = lookup
	delete(s.items, lookup)
	return nil
}
