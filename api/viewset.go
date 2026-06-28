package api

import (
	"context"
	"net/http"
)

// ModelViewSetStore is the persistence boundary used by model viewsets.
type ModelViewSetStore interface {
	List(context.Context, *Request) ([]map[string]any, error)
	Retrieve(context.Context, *Request, string) (map[string]any, error)
	Create(context.Context, *Request, map[string]any) (map[string]any, error)
	Update(context.Context, *Request, string, map[string]any, bool) (map[string]any, error)
	Destroy(context.Context, *Request, string) error
}

// ViewSetAction stores a custom action handler and route metadata.
type ViewSetAction struct {
	Handler View
	Detail  bool
	Methods []string
}

// ModelViewSet provides list, retrieve, create, update, partial update, destroy, and custom actions.
type ModelViewSet struct {
	Store       ModelViewSetStore
	Serializer  *Serializer
	LookupParam string
	View        APIView
	Actions     map[string]ViewSetAction
}

// RegisterAction registers a custom action on the viewset.
func (v *ModelViewSet) RegisterAction(name string, action ViewSetAction) {
	if v.Actions == nil {
		v.Actions = map[string]ViewSetAction{}
	}
	v.Actions[name] = action
}

// AsView returns an API view for a named viewset action.
func (v *ModelViewSet) AsView(action string) View {
	apiView := v.View
	apiView.ParseBody = true
	apiView.Handler = v.dispatch(action)
	return apiView.AsView()
}

func (v *ModelViewSet) dispatch(action string) View {
	switch action {
	case "list":
		return v.List
	case "retrieve":
		return v.Retrieve
	case "create":
		return v.Create
	case "update":
		return v.Update
	case "partial_update":
		return v.PartialUpdate
	case "destroy":
		return v.Destroy
	default:
		if custom, ok := v.Actions[action]; ok && custom.Handler != nil {
			return custom.Handler
		}
		return func(ctx context.Context, request *Request) Response {
			return DefaultExceptionHandler(ctx, request, ErrMethodNotAllowed)
		}
	}
}

// List returns serialized objects.
func (v *ModelViewSet) List(ctx context.Context, request *Request) Response {
	if v.Store == nil {
		return DefaultExceptionHandler(ctx, request, ErrInternal)
	}
	objects, err := v.Store.List(ctx, request)
	if err != nil {
		return DefaultExceptionHandler(ctx, request, err)
	}
	rendered := make([]map[string]any, 0, len(objects))
	for _, object := range objects {
		rendered = append(rendered, v.render(object))
	}
	return JSON(http.StatusOK, rendered)
}

// Retrieve returns one serialized object.
func (v *ModelViewSet) Retrieve(ctx context.Context, request *Request) Response {
	if v.Store == nil {
		return DefaultExceptionHandler(ctx, request, ErrInternal)
	}
	object, err := v.Store.Retrieve(ctx, request, v.lookupValue(request))
	if err != nil {
		return DefaultExceptionHandler(ctx, request, err)
	}
	return JSON(http.StatusOK, v.render(object))
}

// Create validates and creates one object.
func (v *ModelViewSet) Create(ctx context.Context, request *Request) Response {
	if v.Store == nil {
		return DefaultExceptionHandler(ctx, request, ErrInternal)
	}
	data, response, ok := v.validatedBody(request, false)
	if !ok {
		return response
	}
	object, err := v.Store.Create(ctx, request, data)
	if err != nil {
		return DefaultExceptionHandler(ctx, request, err)
	}
	return Created(v.render(object))
}

// Update replaces one object.
func (v *ModelViewSet) Update(ctx context.Context, request *Request) Response {
	return v.update(ctx, request, false)
}

// PartialUpdate updates one object with partial input.
func (v *ModelViewSet) PartialUpdate(ctx context.Context, request *Request) Response {
	return v.update(ctx, request, true)
}

// Destroy deletes one object.
func (v *ModelViewSet) Destroy(ctx context.Context, request *Request) Response {
	if v.Store == nil {
		return DefaultExceptionHandler(ctx, request, ErrInternal)
	}
	if err := v.Store.Destroy(ctx, request, v.lookupValue(request)); err != nil {
		return DefaultExceptionHandler(ctx, request, err)
	}
	return NoContent()
}

func (v *ModelViewSet) update(ctx context.Context, request *Request, partial bool) Response {
	if v.Store == nil {
		return DefaultExceptionHandler(ctx, request, ErrInternal)
	}
	data, response, ok := v.validatedBody(request, partial)
	if !ok {
		return response
	}
	object, err := v.Store.Update(ctx, request, v.lookupValue(request), data, partial)
	if err != nil {
		return DefaultExceptionHandler(ctx, request, err)
	}
	return JSON(http.StatusOK, v.render(object))
}

func (v *ModelViewSet) validatedBody(request *Request, partial bool) (map[string]any, Response, bool) {
	body, ok := request.ParsedBody().(map[string]any)
	if !ok {
		return nil, Error(http.StatusBadRequest, APIError{Code: "invalid_body", Message: "Request body must be an object."}), false
	}
	if v.Serializer == nil {
		return cloneAnyMap(body), Response{}, true
	}
	var (
		data        map[string]any
		fieldErrors map[string][]string
		valid       bool
	)
	if partial {
		data, fieldErrors, valid = v.Serializer.ValidatePartial(body)
	} else {
		data, fieldErrors, valid = v.Serializer.Validate(body)
	}
	if !valid {
		return nil, Error(http.StatusBadRequest, APIError{Code: "validation_error", Message: "Invalid input.", Fields: fieldErrors}), false
	}
	return data, Response{}, true
}

func (v *ModelViewSet) render(object map[string]any) map[string]any {
	if v.Serializer == nil {
		return cloneAnyMap(object)
	}
	return v.Serializer.Render(object)
}

func (v *ModelViewSet) lookupValue(request *Request) string {
	lookupParam := v.LookupParam
	if lookupParam == "" {
		lookupParam = "id"
	}
	if value := request.PathParam(lookupParam); value != "" {
		return value
	}
	if value := request.QueryParam(lookupParam); value != "" {
		return value
	}
	return request.PathParam("pk")
}
