package http

import (
	"context"
	"fmt"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDisplayGenericViewsRenderConfiguredData(t *testing.T) {
	request := NewRequest(httptest.NewRequest("GET", "/", nil))

	template := TemplateView{
		TemplateName: "home.html",
		Context: func(context.Context, *Request) (map[string]any, error) {
			return map[string]any{"title": "Home"}, nil
		},
		Render: func(_ context.Context, _ *Request, templateName string, data map[string]any) Response {
			return Text(nethttp.StatusOK, templateName+":"+data["title"].(string))
		},
	}
	assertViewBody(t, template.AsView(), request, "home.html:Home")

	detail := DetailView{
		GetObject: func(context.Context, *Request) (any, error) {
			return "article-42", nil
		},
		RenderObject: func(_ context.Context, _ *Request, object any) Response {
			return Text(nethttp.StatusOK, object.(string))
		},
	}
	assertViewBody(t, detail.AsView(), request, "article-42")

	list := ListView{
		GetList: func(context.Context, *Request) ([]any, error) {
			return []any{"a", "b"}, nil
		},
		RenderList: func(_ context.Context, _ *Request, values []any) Response {
			return Text(nethttp.StatusOK, fmt.Sprintf("%d", len(values)))
		},
	}
	assertViewBody(t, list.AsView(), request, "2")
}

func TestEditingGenericViewsRunLifecycle(t *testing.T) {
	request := NewRequest(httptest.NewRequest("POST", "/objects/", nil))
	result := FormResult{Valid: true, CleanedData: map[string]any{"name": "created"}}

	create := CreateView{
		Validate: func(context.Context, *Request) (FormResult, error) {
			return result, nil
		},
		Save: func(_ context.Context, _ *Request, form FormResult) (any, error) {
			return form.CleanedData["name"], nil
		},
		Success: func(_ context.Context, _ *Request, object any) Response {
			return Text(nethttp.StatusCreated, object.(string))
		},
	}
	assertViewStatusAndBody(t, create.AsView(), request, nethttp.StatusCreated, "created")

	update := UpdateView{
		GetObject: func(context.Context, *Request) (any, error) {
			return "existing", nil
		},
		Validate: create.Validate,
		Save: func(_ context.Context, _ *Request, object any, _ FormResult) (any, error) {
			return object.(string) + "-updated", nil
		},
		Success: func(_ context.Context, _ *Request, object any) Response {
			return Text(nethttp.StatusOK, object.(string))
		},
	}
	assertViewBody(t, update.AsView(), request, "existing-updated")

	var deleted any
	deleteView := DeleteView{
		GetObject: update.GetObject,
		Delete: func(_ context.Context, _ *Request, object any) error {
			deleted = object
			return nil
		},
		Success: func(context.Context, *Request) Response {
			return Text(nethttp.StatusOK, "deleted")
		},
		Confirm: func(_ context.Context, _ *Request, object any) Response {
			return Text(nethttp.StatusOK, "confirm:"+object.(string))
		},
	}
	assertViewBody(t, deleteView.AsView(), request, "deleted")
	if deleted != "existing" {
		t.Fatalf("deleted object = %v, want existing", deleted)
	}

	getRequest := NewRequest(httptest.NewRequest("GET", "/objects/1/delete/", nil))
	assertViewBody(t, deleteView.AsView(), getRequest, "confirm:existing")
}

func TestDateGenericViewsBuildExpectedPeriods(t *testing.T) {
	now := func() time.Time { return time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC) }
	query := func(_ context.Context, _ *Request, period DatePeriod) ([]any, error) {
		return []any{period.Kind}, nil
	}
	render := func(_ context.Context, _ *Request, data DateArchiveData) Response {
		start := ""
		if !data.Period.Start.IsZero() {
			start = data.Period.Start.Format("2006-01-02")
		}
		return Text(nethttp.StatusOK, string(data.Period.Kind)+":"+start)
	}

	tests := []struct {
		name    string
		view    View
		request *Request
		want    string
	}{
		{
			name:    "archive",
			view:    ArchiveIndexView{Query: query, Render: render}.AsView(),
			request: NewRequest(httptest.NewRequest("GET", "/archive/", nil)),
			want:    "archive:",
		},
		{
			name: "year",
			view: YearArchiveView{Query: query, Render: render}.AsView(),
			request: NewRequest(httptest.NewRequest("GET", "/archive/2026/", nil)).
				WithPathParam("year", "2026"),
			want: "year:2026-01-01",
		},
		{
			name: "month",
			view: MonthArchiveView{Query: query, Render: render}.AsView(),
			request: NewRequest(httptest.NewRequest("GET", "/archive/2026/06/", nil)).
				WithPathParam("year", "2026").WithPathParam("month", "06"),
			want: "month:2026-06-01",
		},
		{
			name: "week",
			view: WeekArchiveView{Query: query, Render: render}.AsView(),
			request: NewRequest(httptest.NewRequest("GET", "/archive/2026/week/26/", nil)).
				WithPathParam("year", "2026").WithPathParam("week", "26"),
			want: "week:2026-06-22",
		},
		{
			name: "day",
			view: DayArchiveView{Query: query, Render: render}.AsView(),
			request: NewRequest(httptest.NewRequest("GET", "/archive/2026/06/27/", nil)).
				WithPathParam("year", "2026").WithPathParam("month", "06").WithPathParam("day", "27"),
			want: "day:2026-06-27",
		},
		{
			name:    "today",
			view:    TodayArchiveView{Now: now, Query: query, Render: render}.AsView(),
			request: NewRequest(httptest.NewRequest("GET", "/archive/today/", nil)),
			want:    "today:2026-06-27",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertViewBody(t, test.view, test.request, test.want)
		})
	}

	detail := DateDetailView{
		GetObject: func(_ context.Context, _ *Request, period DatePeriod) (any, error) {
			return period.Start.Format("2006-01-02"), nil
		},
		RenderObject: func(_ context.Context, _ *Request, object any, period DatePeriod) Response {
			return Text(nethttp.StatusOK, string(period.Kind)+":"+object.(string))
		},
	}
	request := NewRequest(httptest.NewRequest("GET", "/archive/2026/06/27/item/", nil)).
		WithPathParam("year", "2026").WithPathParam("month", "06").WithPathParam("day", "27")
	assertViewBody(t, detail.AsView(), request, "date-detail:2026-06-27")
}

func assertViewBody(t *testing.T, view View, request *Request, want string) {
	t.Helper()
	assertViewStatusAndBody(t, view, request, nethttp.StatusOK, want)
}

func assertViewStatusAndBody(t *testing.T, view View, request *Request, status int, want string) {
	t.Helper()

	response := view(request.Context(), request)
	recorder := httptest.NewRecorder()
	if err := response.Write(recorder); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if recorder.Code != status {
		t.Fatalf("status = %d, want %d", recorder.Code, status)
	}
	if recorder.Body.String() != want {
		t.Fatalf("body = %q, want %q", recorder.Body.String(), want)
	}
}
