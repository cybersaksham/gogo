package api

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"unicode"
)

// StoredUpload is the storage result for one uploaded file.
type StoredUpload struct {
	Name        string
	Size        int64
	ContentType string
	Location    string
}

// UploadStorage stores validated uploads.
type UploadStorage interface {
	SaveUpload(context.Context, UploadedFile) (StoredUpload, error)
}

// UploadConfig configures upload validation and storage.
type UploadConfig struct {
	FieldName         string
	MaxSize           int64
	AllowedExtensions []string
	ImageValidator    func(UploadedFile) error
	Storage           UploadStorage
}

// UploadHandler validates and stores multipart or streamed uploads.
type UploadHandler struct {
	Config UploadConfig
}

// MemoryUploadStorage stores uploads in memory for tests, examples, and bootstrapping.
type MemoryUploadStorage struct {
	Files    []StoredUpload
	Contents map[string][]byte
}

// NewMemoryUploadStorage creates an in-memory upload storage.
func NewMemoryUploadStorage() *MemoryUploadStorage {
	return &MemoryUploadStorage{Contents: map[string][]byte{}}
}

// SaveUpload stores one upload.
func (s *MemoryUploadStorage) SaveUpload(_ context.Context, file UploadedFile) (StoredUpload, error) {
	if s.Contents == nil {
		s.Contents = map[string][]byte{}
	}
	stored := StoredUpload{
		Name:        file.Filename,
		Size:        uploadSize(file),
		ContentType: http.DetectContentType(file.Content),
		Location:    file.Filename,
	}
	s.Files = append(s.Files, stored)
	s.Contents[file.Filename] = append([]byte(nil), file.Content...)
	return stored, nil
}

// HandleMultipart parses, validates, and stores multipart uploads.
func (h UploadHandler) HandleMultipart(ctx context.Context, request *Request) ([]StoredUpload, error) {
	body := request.ParsedBody()
	if body == nil {
		parsed, err := DefaultParserRegistry().Parse(request.Raw(), multipartParseLimit(h.Config.MaxSize))
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUpload, err)
		}
		body = parsed
		request.WithParsedBody(parsed)
	}
	multipartBody, ok := body.(MultipartBody)
	if !ok {
		return nil, fmt.Errorf("%w: request body is not multipart", ErrUpload)
	}
	files := multipartFiles(multipartBody, h.Config.FieldName)
	stored := make([]StoredUpload, 0, len(files))
	for _, file := range files {
		value, err := h.SaveUploadedFile(ctx, file)
		if err != nil {
			return nil, err
		}
		stored = append(stored, value)
	}
	return stored, nil
}

// HandleStream validates and stores a raw request body as one file.
func (h UploadHandler) HandleStream(ctx context.Context, request *Request, filename string) (StoredUpload, error) {
	if filename == "" {
		filename = request.QueryParam("filename")
	}
	if filename == "" {
		filename = request.Raw().Header.Get("X-Filename")
	}
	content, err := readLimited(request.Raw().Body, h.Config.MaxSize)
	if err != nil {
		return StoredUpload{}, fmt.Errorf("%w: %w", ErrUpload, err)
	}
	return h.SaveUploadedFile(ctx, UploadedFile{Filename: filename, Size: int64(len(content)), Content: content})
}

// SaveUploadedFile validates and stores one parsed upload.
func (h UploadHandler) SaveUploadedFile(ctx context.Context, file UploadedFile) (StoredUpload, error) {
	name, err := normalizedUploadName(file.Filename)
	if err != nil {
		return StoredUpload{}, err
	}
	file.Filename = name
	size := uploadSize(file)
	if h.Config.MaxSize > 0 && size > h.Config.MaxSize {
		return StoredUpload{}, fmt.Errorf("%w: file exceeds size limit", ErrUpload)
	}
	if err := validateUploadExtension(name, h.Config.AllowedExtensions); err != nil {
		return StoredUpload{}, err
	}
	if h.Config.ImageValidator != nil {
		if err := h.Config.ImageValidator(file); err != nil {
			return StoredUpload{}, fmt.Errorf("%w: %w", ErrUpload, err)
		}
	}
	if h.Config.Storage == nil {
		return StoredUpload{}, fmt.Errorf("%w: storage backend is required", ErrUpload)
	}
	stored, err := h.Config.Storage.SaveUpload(ctx, file)
	if err != nil {
		return StoredUpload{}, fmt.Errorf("%w: %w", ErrUpload, err)
	}
	return stored, nil
}

// ImageSignatureValidator accepts PNG, JPEG, and GIF upload signatures.
func ImageSignatureValidator(file UploadedFile) error {
	content := file.Content
	switch {
	case bytes.HasPrefix(content, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return nil
	case bytes.HasPrefix(content, []byte{0xff, 0xd8, 0xff}):
		return nil
	case bytes.HasPrefix(content, []byte("GIF87a")), bytes.HasPrefix(content, []byte("GIF89a")):
		return nil
	default:
		return fmt.Errorf("%w: invalid image signature", ErrUpload)
	}
}

func multipartFiles(body MultipartBody, fieldName string) []UploadedFile {
	if fieldName != "" {
		return append([]UploadedFile(nil), body.Files[fieldName]...)
	}
	var files []UploadedFile
	for _, values := range body.Files {
		files = append(files, values...)
	}
	return files
}

func multipartParseLimit(maxSize int64) int64 {
	if maxSize <= 0 {
		return DefaultBodyLimit
	}
	return maxSize + (1 << 20)
}

func uploadSize(file UploadedFile) int64 {
	if file.Size > 0 {
		return file.Size
	}
	return int64(len(file.Content))
}

func normalizedUploadName(filename string) (string, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "", fmt.Errorf("%w: filename is required", ErrUpload)
	}
	if strings.Contains(filename, "\x00") || strings.Contains(filename, "\\") || filepath.Base(filename) != filename {
		return "", fmt.Errorf("%w: unsafe filename", ErrUpload)
	}
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	base = sanitizeUploadBase(base)
	if base == "" {
		base = "upload"
	}
	return base + ext, nil
}

func sanitizeUploadBase(value string) string {
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range value {
		allowed := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.'
		if allowed {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(builder.String(), "._-")
}

func validateUploadExtension(filename string, allowed []string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return fmt.Errorf("%w: file extension is required", ErrUpload)
	}
	if dangerousUploadExtension(ext) {
		return fmt.Errorf("%w: executable file extensions are not allowed", ErrUpload)
	}
	if len(allowed) == 0 {
		return nil
	}
	allowedSet := map[string]struct{}{}
	for _, value := range allowed {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if !strings.HasPrefix(value, ".") {
			value = "." + value
		}
		allowedSet[value] = struct{}{}
	}
	if _, ok := allowedSet[ext]; !ok {
		return fmt.Errorf("%w: file extension is not allowed", ErrUpload)
	}
	return nil
}

func dangerousUploadExtension(ext string) bool {
	switch ext {
	case ".bat", ".cmd", ".com", ".exe", ".jar", ".js", ".msi", ".php", ".ps1", ".scr", ".sh":
		return true
	default:
		return false
	}
}
