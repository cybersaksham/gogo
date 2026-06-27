package http

import (
	"context"
	"fmt"
	nethttp "net/http"
	"strconv"
	"time"
)

// BaseView is the smallest generic view wrapper.
type BaseView struct {
	Handler View
	Methods []string
}

// AsView returns the executable framework view.
func (v BaseView) AsView() View {
	handler := v.Handler
	if handler == nil {
		handler = func(context.Context, *Request) Response {
			return NoContent()
		}
	}
	if len(v.Methods) > 0 {
		return RequireHTTPMethods(v.Methods...)(handler)
	}
	return handler
}

// TemplateView renders a named template with context data.
type TemplateView struct {
	TemplateName string
	Context      func(context.Context, *Request) (map[string]any, error)
	Render       func(context.Context, *Request, string, map[string]any) Response
}

// AsView returns a view for the template renderer.
func (v TemplateView) AsView() View {
	return RequireSafeMethods(func(ctx context.Context, request *Request) Response {
		data := map[string]any{}
		if v.Context != nil {
			var err error
			data, err = v.Context(ctx, request)
			if err != nil {
				return internalError()
			}
		}
		if v.Render != nil {
			return v.Render(ctx, request, v.TemplateName, data)
		}

		response := JSON(nethttp.StatusOK, data)
		response.Header().Set("X-Template-Name", v.TemplateName)
		return response
	})
}

// RedirectView redirects to a configured URL.
type RedirectView struct {
	URL         string
	Permanent   bool
	QueryString bool
	GetURL      func(context.Context, *Request) (string, error)
}

// AsView returns a redirecting view.
func (v RedirectView) AsView() View {
	return func(ctx context.Context, request *Request) Response {
		target := v.URL
		if v.GetURL != nil {
			url, err := v.GetURL(ctx, request)
			if err != nil {
				return internalError()
			}
			target = url
		}
		if v.QueryString && request.Raw().URL.RawQuery != "" {
			separator := "?"
			if containsQuery(target) {
				separator = "&"
			}
			target += separator + request.Raw().URL.RawQuery
		}

		status := nethttp.StatusFound
		if v.Permanent {
			status = nethttp.StatusMovedPermanently
		}
		response := Text(status, "")
		response.Header().Set("Location", target)
		return response
	}
}

// DetailView renders one object.
type DetailView struct {
	GetObject    func(context.Context, *Request) (any, error)
	RenderObject func(context.Context, *Request, any) Response
}

// AsView returns a detail view.
func (v DetailView) AsView() View {
	return RequireSafeMethods(func(ctx context.Context, request *Request) Response {
		if v.GetObject == nil {
			return internalError()
		}
		object, err := v.GetObject(ctx, request)
		if err != nil {
			return internalError()
		}
		if v.RenderObject != nil {
			return v.RenderObject(ctx, request, object)
		}
		return JSON(nethttp.StatusOK, object)
	})
}

// ListView renders a list of objects.
type ListView struct {
	GetList    func(context.Context, *Request) ([]any, error)
	RenderList func(context.Context, *Request, []any) Response
}

// AsView returns a list view.
func (v ListView) AsView() View {
	return RequireSafeMethods(func(ctx context.Context, request *Request) Response {
		if v.GetList == nil {
			return internalError()
		}
		values, err := v.GetList(ctx, request)
		if err != nil {
			return internalError()
		}
		if v.RenderList != nil {
			return v.RenderList(ctx, request, values)
		}
		return JSON(nethttp.StatusOK, values)
	})
}

// FormResult is the result of validating a request body.
type FormResult struct {
	Valid       bool
	CleanedData map[string]any
	Errors      map[string]string
}

// FormView handles form display and submission.
type FormView struct {
	Validate func(context.Context, *Request) (FormResult, error)
	Render   func(context.Context, *Request, FormResult) Response
	Success  func(context.Context, *Request, FormResult) Response
}

// AsView returns a form view.
func (v FormView) AsView() View {
	return func(ctx context.Context, request *Request) Response {
		if request.Method() == nethttp.MethodGet {
			return v.render(ctx, request, FormResult{})
		}
		if !contains([]string{nethttp.MethodPost, nethttp.MethodPut, nethttp.MethodPatch}, request.Method()) {
			return methodNotAllowed([]string{nethttp.MethodGet, nethttp.MethodPost, nethttp.MethodPut, nethttp.MethodPatch})
		}

		form, err := v.validate(ctx, request)
		if err != nil {
			return internalError()
		}
		if !form.Valid {
			return v.render(ctx, request, form)
		}
		if v.Success != nil {
			return v.Success(ctx, request, form)
		}
		return NoContent()
	}
}

func (v FormView) validate(ctx context.Context, request *Request) (FormResult, error) {
	if v.Validate == nil {
		return FormResult{Valid: true}, nil
	}
	return v.Validate(ctx, request)
}

func (v FormView) render(ctx context.Context, request *Request, result FormResult) Response {
	if v.Render != nil {
		return v.Render(ctx, request, result)
	}
	return JSON(nethttp.StatusOK, result)
}

// CreateView validates form data and creates an object.
type CreateView struct {
	Validate func(context.Context, *Request) (FormResult, error)
	Save     func(context.Context, *Request, FormResult) (any, error)
	Render   func(context.Context, *Request, FormResult) Response
	Success  func(context.Context, *Request, any) Response
}

// AsView returns a create view.
func (v CreateView) AsView() View {
	return RequirePOST(func(ctx context.Context, request *Request) Response {
		form, err := v.validate(ctx, request)
		if err != nil {
			return internalError()
		}
		if !form.Valid {
			return v.render(ctx, request, form)
		}
		if v.Save == nil {
			return internalError()
		}
		object, err := v.Save(ctx, request, form)
		if err != nil {
			return internalError()
		}
		if v.Success != nil {
			return v.Success(ctx, request, object)
		}
		return JSON(nethttp.StatusCreated, object)
	})
}

func (v CreateView) validate(ctx context.Context, request *Request) (FormResult, error) {
	if v.Validate == nil {
		return FormResult{Valid: true}, nil
	}
	return v.Validate(ctx, request)
}

func (v CreateView) render(ctx context.Context, request *Request, result FormResult) Response {
	if v.Render != nil {
		return v.Render(ctx, request, result)
	}
	return JSON(nethttp.StatusBadRequest, result)
}

// UpdateView validates form data and updates an existing object.
type UpdateView struct {
	GetObject func(context.Context, *Request) (any, error)
	Validate  func(context.Context, *Request) (FormResult, error)
	Save      func(context.Context, *Request, any, FormResult) (any, error)
	Render    func(context.Context, *Request, FormResult) Response
	Success   func(context.Context, *Request, any) Response
}

// AsView returns an update view.
func (v UpdateView) AsView() View {
	return RequireHTTPMethods(nethttp.MethodPost, nethttp.MethodPut, nethttp.MethodPatch)(func(ctx context.Context, request *Request) Response {
		object, err := v.getObject(ctx, request)
		if err != nil {
			return internalError()
		}
		form, err := v.validate(ctx, request)
		if err != nil {
			return internalError()
		}
		if !form.Valid {
			return v.render(ctx, request, form)
		}
		if v.Save == nil {
			return internalError()
		}
		updated, err := v.Save(ctx, request, object, form)
		if err != nil {
			return internalError()
		}
		if v.Success != nil {
			return v.Success(ctx, request, updated)
		}
		return JSON(nethttp.StatusOK, updated)
	})
}

func (v UpdateView) getObject(ctx context.Context, request *Request) (any, error) {
	if v.GetObject == nil {
		return nil, fmt.Errorf("missing object getter")
	}
	return v.GetObject(ctx, request)
}

func (v UpdateView) validate(ctx context.Context, request *Request) (FormResult, error) {
	if v.Validate == nil {
		return FormResult{Valid: true}, nil
	}
	return v.Validate(ctx, request)
}

func (v UpdateView) render(ctx context.Context, request *Request, result FormResult) Response {
	if v.Render != nil {
		return v.Render(ctx, request, result)
	}
	return JSON(nethttp.StatusBadRequest, result)
}

// DeleteView confirms and deletes an existing object.
type DeleteView struct {
	GetObject func(context.Context, *Request) (any, error)
	Delete    func(context.Context, *Request, any) error
	Confirm   func(context.Context, *Request, any) Response
	Success   func(context.Context, *Request) Response
}

// AsView returns a delete view.
func (v DeleteView) AsView() View {
	return func(ctx context.Context, request *Request) Response {
		object, err := v.getObject(ctx, request)
		if err != nil {
			return internalError()
		}
		switch request.Method() {
		case nethttp.MethodGet, nethttp.MethodHead:
			if v.Confirm != nil {
				return v.Confirm(ctx, request, object)
			}
			return JSON(nethttp.StatusOK, object)
		case nethttp.MethodPost, nethttp.MethodDelete:
			if v.Delete == nil {
				return internalError()
			}
			if err := v.Delete(ctx, request, object); err != nil {
				return internalError()
			}
			if v.Success != nil {
				return v.Success(ctx, request)
			}
			return NoContent()
		default:
			return methodNotAllowed([]string{nethttp.MethodGet, nethttp.MethodPost, nethttp.MethodDelete})
		}
	}
}

func (v DeleteView) getObject(ctx context.Context, request *Request) (any, error) {
	if v.GetObject == nil {
		return nil, fmt.Errorf("missing object getter")
	}
	return v.GetObject(ctx, request)
}

// DatePeriodKind identifies the active archive view kind.
type DatePeriodKind string

const (
	DatePeriodArchive    DatePeriodKind = "archive"
	DatePeriodYear       DatePeriodKind = "year"
	DatePeriodMonth      DatePeriodKind = "month"
	DatePeriodWeek       DatePeriodKind = "week"
	DatePeriodDay        DatePeriodKind = "day"
	DatePeriodToday      DatePeriodKind = "today"
	DatePeriodDateDetail DatePeriodKind = "date-detail"
)

// DatePeriod represents an inclusive-exclusive date range.
type DatePeriod struct {
	Kind  DatePeriodKind
	Start time.Time
	End   time.Time
}

// DateArchiveData contains archive render data.
type DateArchiveData struct {
	Period  DatePeriod
	Objects []any
}

// ArchiveIndexView renders archive root data.
type ArchiveIndexView struct {
	Query  func(context.Context, *Request, DatePeriod) ([]any, error)
	Render func(context.Context, *Request, DateArchiveData) Response
}

// AsView returns an archive index view.
func (v ArchiveIndexView) AsView() View {
	return v.archiveView(DatePeriod{Kind: DatePeriodArchive})
}

// YearArchiveView renders one year.
type YearArchiveView struct {
	Query  func(context.Context, *Request, DatePeriod) ([]any, error)
	Render func(context.Context, *Request, DateArchiveData) Response
}

// AsView returns a year archive view.
func (v YearArchiveView) AsView() View {
	return func(ctx context.Context, request *Request) Response {
		year, ok := pathParamInt(request, "year")
		if !ok {
			return Text(nethttp.StatusBadRequest, "Bad Request")
		}
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		return archiveResponse(ctx, request, v.Query, v.Render, DatePeriod{Kind: DatePeriodYear, Start: start, End: start.AddDate(1, 0, 0)})
	}
}

// MonthArchiveView renders one month.
type MonthArchiveView struct {
	Query  func(context.Context, *Request, DatePeriod) ([]any, error)
	Render func(context.Context, *Request, DateArchiveData) Response
}

// AsView returns a month archive view.
func (v MonthArchiveView) AsView() View {
	return func(ctx context.Context, request *Request) Response {
		year, month, ok := yearMonth(request)
		if !ok {
			return Text(nethttp.StatusBadRequest, "Bad Request")
		}
		start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		return archiveResponse(ctx, request, v.Query, v.Render, DatePeriod{Kind: DatePeriodMonth, Start: start, End: start.AddDate(0, 1, 0)})
	}
}

// WeekArchiveView renders one ISO week.
type WeekArchiveView struct {
	Query  func(context.Context, *Request, DatePeriod) ([]any, error)
	Render func(context.Context, *Request, DateArchiveData) Response
}

// AsView returns a week archive view.
func (v WeekArchiveView) AsView() View {
	return func(ctx context.Context, request *Request) Response {
		year, ok := pathParamInt(request, "year")
		if !ok {
			return Text(nethttp.StatusBadRequest, "Bad Request")
		}
		week, ok := pathParamInt(request, "week")
		if !ok {
			return Text(nethttp.StatusBadRequest, "Bad Request")
		}
		start := isoWeekStart(year, week)
		return archiveResponse(ctx, request, v.Query, v.Render, DatePeriod{Kind: DatePeriodWeek, Start: start, End: start.AddDate(0, 0, 7)})
	}
}

// DayArchiveView renders one day.
type DayArchiveView struct {
	Query  func(context.Context, *Request, DatePeriod) ([]any, error)
	Render func(context.Context, *Request, DateArchiveData) Response
}

// AsView returns a day archive view.
func (v DayArchiveView) AsView() View {
	return func(ctx context.Context, request *Request) Response {
		date, ok := ymd(request)
		if !ok {
			return Text(nethttp.StatusBadRequest, "Bad Request")
		}
		return archiveResponse(ctx, request, v.Query, v.Render, DatePeriod{Kind: DatePeriodDay, Start: date, End: date.AddDate(0, 0, 1)})
	}
}

// TodayArchiveView renders today's archive.
type TodayArchiveView struct {
	Now    func() time.Time
	Query  func(context.Context, *Request, DatePeriod) ([]any, error)
	Render func(context.Context, *Request, DateArchiveData) Response
}

// AsView returns a today archive view.
func (v TodayArchiveView) AsView() View {
	return func(ctx context.Context, request *Request) Response {
		now := time.Now().UTC()
		if v.Now != nil {
			now = v.Now().UTC()
		}
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return archiveResponse(ctx, request, v.Query, v.Render, DatePeriod{Kind: DatePeriodToday, Start: start, End: start.AddDate(0, 0, 1)})
	}
}

// DateDetailView renders one object within a date period.
type DateDetailView struct {
	GetObject    func(context.Context, *Request, DatePeriod) (any, error)
	RenderObject func(context.Context, *Request, any, DatePeriod) Response
}

// AsView returns a date detail view.
func (v DateDetailView) AsView() View {
	return func(ctx context.Context, request *Request) Response {
		date, ok := ymd(request)
		if !ok {
			return Text(nethttp.StatusBadRequest, "Bad Request")
		}
		period := DatePeriod{Kind: DatePeriodDateDetail, Start: date, End: date.AddDate(0, 0, 1)}
		if v.GetObject == nil {
			return internalError()
		}
		object, err := v.GetObject(ctx, request, period)
		if err != nil {
			return internalError()
		}
		if v.RenderObject != nil {
			return v.RenderObject(ctx, request, object, period)
		}
		return JSON(nethttp.StatusOK, object)
	}
}

func (v ArchiveIndexView) archiveView(period DatePeriod) View {
	return func(ctx context.Context, request *Request) Response {
		return archiveResponse(ctx, request, v.Query, v.Render, period)
	}
}

func archiveResponse(ctx context.Context, request *Request, query func(context.Context, *Request, DatePeriod) ([]any, error), render func(context.Context, *Request, DateArchiveData) Response, period DatePeriod) Response {
	objects := []any{}
	if query != nil {
		var err error
		objects, err = query(ctx, request, period)
		if err != nil {
			return internalError()
		}
	}
	data := DateArchiveData{Period: period, Objects: objects}
	if render != nil {
		return render(ctx, request, data)
	}
	return JSON(nethttp.StatusOK, data)
}

func ymd(request *Request) (time.Time, bool) {
	year, month, ok := yearMonth(request)
	if !ok {
		return time.Time{}, false
	}
	day, ok := pathParamInt(request, "day")
	if !ok {
		return time.Time{}, false
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), true
}

func yearMonth(request *Request) (int, int, bool) {
	year, ok := pathParamInt(request, "year")
	if !ok {
		return 0, 0, false
	}
	month, ok := pathParamInt(request, "month")
	if !ok || month < 1 || month > 12 {
		return 0, 0, false
	}
	return year, month, true
}

func pathParamInt(request *Request, name string) (int, bool) {
	value, err := strconv.Atoi(request.PathParam(name))
	return value, err == nil
}

func isoWeekStart(year, week int) time.Time {
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	weekday := int(jan4.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return jan4.AddDate(0, 0, -(weekday-1)).AddDate(0, 0, (week-1)*7)
}

func containsQuery(target string) bool {
	for _, char := range target {
		if char == '?' {
			return true
		}
		if char == '#' {
			return false
		}
	}
	return false
}

func internalError() Response {
	return Text(nethttp.StatusInternalServerError, "Internal Server Error")
}
