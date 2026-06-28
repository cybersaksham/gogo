package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

const DefaultBodyLimit int64 = 32 << 20

// View handles one API request and returns an API response.
type View func(context.Context, *Request) Response

// RequestInitializer prepares a request before authentication and parsing.
type RequestInitializer func(context.Context, *Request) (*Request, error)

// RequestHook validates one request lifecycle stage.
type RequestHook func(context.Context, *Request) error

// ExceptionHandler converts lifecycle errors into API responses.
type ExceptionHandler func(context.Context, *Request, error) Response

// ResponseFinalizer runs after handler or exception responses are produced.
type ResponseFinalizer func(context.Context, *Request, Response) Response

// APIView is the struct-based API view with Django REST Framework-style hooks.
type APIView struct {
	Handler             View
	ParserRegistry      ParserRegistry
	ParseBody           bool
	BodyLimit           int64
	InitializeRequest   RequestInitializer
	CheckAuthentication RequestHook
	CheckPermissions    RequestHook
	CheckThrottles      RequestHook
	HandleException     ExceptionHandler
	FinalizeResponse    ResponseFinalizer
}

// FunctionView adapts a function API view through the default lifecycle.
func FunctionView(handler View) View {
	return APIView{Handler: handler}.AsView()
}

// AsView composes lifecycle hooks around the struct-based view handler.
func (v APIView) AsView() View {
	return func(ctx context.Context, request *Request) (response Response) {
		if request == nil {
			return v.exception(ctx, request, ErrInternal)
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				response = v.finalize(ctx, request, v.exception(ctx, request, fmt.Errorf("%w: %v", ErrInternal, recovered)))
			}
		}()

		initialized, err := v.initialize(ctx, request)
		if err != nil {
			return v.finalize(ctx, request, v.exception(ctx, request, err))
		}
		request = initialized
		if err := runRequestHook(ctx, request, v.CheckAuthentication); err != nil {
			return v.finalize(ctx, request, v.exception(ctx, request, err))
		}
		if err := runRequestHook(ctx, request, v.CheckPermissions); err != nil {
			return v.finalize(ctx, request, v.exception(ctx, request, err))
		}
		if err := runRequestHook(ctx, request, v.CheckThrottles); err != nil {
			return v.finalize(ctx, request, v.exception(ctx, request, err))
		}
		if err := v.parseBody(request); err != nil {
			return v.finalize(ctx, request, v.exception(ctx, request, err))
		}
		if v.Handler == nil {
			return v.finalize(ctx, request, v.exception(ctx, request, ErrMethodNotAllowed))
		}
		return v.finalize(ctx, request, v.Handler(ctx, request))
	}
}

// DefaultExceptionHandler returns safe, normalized API errors.
func DefaultExceptionHandler(_ context.Context, _ *Request, err error) Response {
	status, code, message := apiErrorStatus(err)
	return Error(status, APIError{Code: code, Message: message})
}

func (v APIView) initialize(ctx context.Context, request *Request) (*Request, error) {
	if v.InitializeRequest == nil {
		return request, nil
	}
	return v.InitializeRequest(ctx, request)
}

func (v APIView) exception(ctx context.Context, request *Request, err error) Response {
	if v.HandleException != nil {
		return v.HandleException(ctx, request, err)
	}
	return DefaultExceptionHandler(ctx, request, err)
}

func (v APIView) finalize(ctx context.Context, request *Request, response Response) Response {
	if v.FinalizeResponse == nil {
		return response
	}
	return v.FinalizeResponse(ctx, request, response)
}

func (v APIView) parseBody(request *Request) error {
	if !v.ParseBody || request.ParsedBody() != nil || request.raw == nil || request.raw.Body == nil || request.raw.Body == http.NoBody || request.raw.ContentLength == 0 {
		return nil
	}
	registry := v.ParserRegistry
	if registry.parsers == nil {
		registry = DefaultParserRegistry()
	}
	limit := v.BodyLimit
	if limit <= 0 {
		limit = DefaultBodyLimit
	}
	body, err := registry.Parse(request.Raw(), limit)
	if err != nil {
		return err
	}
	request.WithParsedBody(body)
	return nil
}

func runRequestHook(ctx context.Context, request *Request, hook RequestHook) error {
	if hook == nil {
		return nil
	}
	return hook(ctx, request)
}

func apiErrorStatus(err error) (int, string, string) {
	switch {
	case errors.Is(err, ErrAuthenticationFailed):
		return http.StatusUnauthorized, "authentication_failed", "Authentication credentials were not provided or invalid."
	case errors.Is(err, ErrPermissionDenied):
		return http.StatusForbidden, "permission_denied", "Permission denied."
	case errors.Is(err, ErrThrottled):
		return http.StatusTooManyRequests, "throttled", "Request was throttled."
	case errors.Is(err, ErrUnsupportedMediaType):
		return http.StatusUnsupportedMediaType, "unsupported_media_type", "Unsupported media type."
	case errors.Is(err, ErrBodyTooLarge):
		return http.StatusRequestEntityTooLarge, "body_too_large", "Request body is too large."
	case errors.Is(err, ErrParse):
		return http.StatusBadRequest, "parse_error", "Request body could not be parsed."
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest, "validation_error", "Invalid input."
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, "not_found", "Not found."
	case errors.Is(err, ErrMethodNotAllowed):
		return http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed."
	default:
		return http.StatusInternalServerError, "internal_error", "Internal server error."
	}
}
