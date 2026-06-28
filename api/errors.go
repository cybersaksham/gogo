package api

import "errors"

var (
	ErrUnsupportedMediaType    = errors.New("unsupported media type")
	ErrParse                   = errors.New("parse error")
	ErrBodyTooLarge            = errors.New("body too large")
	ErrNotAcceptable           = errors.New("not acceptable")
	ErrValidation              = errors.New("validation error")
	ErrInvalidSerializerConfig = errors.New("invalid serializer config")
	ErrAuthenticationFailed    = errors.New("authentication failed")
	ErrPermissionDenied        = errors.New("permission denied")
	ErrThrottled               = errors.New("request throttled")
	ErrNotFound                = errors.New("not found")
	ErrMethodNotAllowed        = errors.New("method not allowed")
	ErrInternal                = errors.New("internal error")
	ErrRouteConflict           = errors.New("route conflict")
	ErrReverse                 = errors.New("reverse error")
	ErrPagination              = errors.New("pagination error")
	ErrFilter                  = errors.New("filter error")
	ErrUpload                  = errors.New("upload error")
	ErrVersion                 = errors.New("version error")
)
