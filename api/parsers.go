package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

// UploadedFile stores parsed multipart file metadata.
type UploadedFile struct {
	Filename string
	Size     int64
	Content  []byte
}

// MultipartBody stores multipart values and files.
type MultipartBody struct {
	Values map[string][]string
	Files  map[string][]UploadedFile
}

// Parser parses an HTTP request body.
type Parser func(*http.Request, int64) (any, error)

// ParserRegistry maps media types to parsers.
type ParserRegistry struct {
	parsers map[string]Parser
}

// DefaultParserRegistry returns built-in parsers.
func DefaultParserRegistry() ParserRegistry {
	return ParserRegistry{parsers: map[string]Parser{
		"application/json":                  parseJSON,
		"application/x-www-form-urlencoded": parseForm,
		"multipart/form-data":               parseMultipart,
		"text/plain":                        parseText,
		"application/octet-stream":          parseRaw,
	}}
}

// Parse parses a request using Content-Type and a byte limit.
func (r ParserRegistry) Parse(request *http.Request, limit int64) (any, error) {
	mediaType := request.Header.Get("Content-Type")
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	base, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedMediaType, err)
	}
	parser, ok := r.parsers[strings.ToLower(base)]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedMediaType, base)
	}
	return parser(request, limit)
}

func parseJSON(request *http.Request, limit int64) (any, error) {
	body, err := readLimited(request.Body, limit)
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(body, &value); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}
	return value, nil
}

func parseForm(request *http.Request, limit int64) (any, error) {
	body, err := readLimited(request.Body, limit)
	if err != nil {
		return nil, err
	}
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}
	return map[string][]string(values), nil
}

func parseMultipart(request *http.Request, limit int64) (any, error) {
	if limit <= 0 {
		limit = 32 << 20
	}
	request.Body = http.MaxBytesReader(nil, request.Body, limit)
	if err := request.ParseMultipartForm(limit); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}
	body := MultipartBody{Values: map[string][]string{}, Files: map[string][]UploadedFile{}}
	if request.MultipartForm != nil {
		for key, values := range request.MultipartForm.Value {
			body.Values[key] = append([]string(nil), values...)
		}
		for key, headers := range request.MultipartForm.File {
			for _, header := range headers {
				file, err := header.Open()
				if err != nil {
					return nil, err
				}
				content, err := io.ReadAll(file)
				_ = file.Close()
				if err != nil {
					return nil, err
				}
				body.Files[key] = append(body.Files[key], UploadedFile{Filename: header.Filename, Size: header.Size, Content: content})
			}
		}
	}
	return body, nil
}

func parseText(request *http.Request, limit int64) (any, error) {
	body, err := readLimited(request.Body, limit)
	if err != nil {
		return nil, err
	}
	return string(body), nil
}

func parseRaw(request *http.Request, limit int64) (any, error) {
	return readLimited(request.Body, limit)
}

func readLimited(reader io.Reader, limit int64) ([]byte, error) {
	if limit <= 0 {
		limit = 32 << 20
	}
	limited := io.LimitReader(reader, limit+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, ErrBodyTooLarge
	}
	return body, nil
}
