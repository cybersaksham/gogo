package api

import "errors"

var (
	ErrUnsupportedMediaType    = errors.New("unsupported media type")
	ErrParse                   = errors.New("parse error")
	ErrBodyTooLarge            = errors.New("body too large")
	ErrNotAcceptable           = errors.New("not acceptable")
	ErrValidation              = errors.New("validation error")
	ErrInvalidSerializerConfig = errors.New("invalid serializer config")
)
