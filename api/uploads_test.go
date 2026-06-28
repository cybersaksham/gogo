package api

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMultipartUploadValidatesImageAndStoresFile(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "avatar.png")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	part.Write([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'})
	writer.Close()

	raw := httptest.NewRequest(http.MethodPost, "/upload/", &body)
	raw.Header.Set("Content-Type", writer.FormDataContentType())
	storage := NewMemoryUploadStorage()
	handler := UploadHandler{Config: UploadConfig{
		FieldName:         "file",
		MaxSize:           32,
		AllowedExtensions: []string{".png"},
		ImageValidator:    ImageSignatureValidator,
		Storage:           storage,
	}}

	files, err := handler.HandleMultipart(context.Background(), NewRequest(raw))
	if err != nil {
		t.Fatalf("HandleMultipart() error = %v", err)
	}
	if len(files) != 1 || files[0].Name != "avatar.png" || len(storage.Files) != 1 {
		t.Fatalf("stored files = %#v storage=%#v", files, storage.Files)
	}
}

func TestStreamUploadNormalizesNameAndStoresFile(t *testing.T) {
	raw := httptest.NewRequest(http.MethodPut, "/upload/?filename=Report%20Final.txt", bytes.NewReader([]byte("report")))
	storage := NewMemoryUploadStorage()
	handler := UploadHandler{Config: UploadConfig{MaxSize: 10, Storage: storage}}

	file, err := handler.HandleStream(context.Background(), NewRequest(raw), "")
	if err != nil {
		t.Fatalf("HandleStream() error = %v", err)
	}
	if file.Name != "Report_Final.txt" || len(storage.Files) != 1 {
		t.Fatalf("file = %#v storage=%#v", file, storage.Files)
	}
}

func TestUploadRejectsSizeExtensionAndPathTraversal(t *testing.T) {
	handler := UploadHandler{Config: UploadConfig{MaxSize: 3, AllowedExtensions: []string{".txt"}, Storage: NewMemoryUploadStorage()}}
	_, err := handler.SaveUploadedFile(context.Background(), UploadedFile{Filename: "safe.txt", Content: []byte("toolarge")})
	if !errors.Is(err, ErrUpload) {
		t.Fatalf("size error = %v, want ErrUpload", err)
	}

	handler.Config.MaxSize = 100
	_, err = handler.SaveUploadedFile(context.Background(), UploadedFile{Filename: "script.exe", Content: []byte("x")})
	if !errors.Is(err, ErrUpload) {
		t.Fatalf("extension error = %v, want ErrUpload", err)
	}

	_, err = handler.SaveUploadedFile(context.Background(), UploadedFile{Filename: "../safe.txt", Content: []byte("x")})
	if !errors.Is(err, ErrUpload) {
		t.Fatalf("path traversal error = %v, want ErrUpload", err)
	}
}

func TestUploadPropagatesStorageErrors(t *testing.T) {
	storageErr := errors.New("storage unavailable")
	handler := UploadHandler{Config: UploadConfig{Storage: failingUploadStorage{err: storageErr}}}

	_, err := handler.SaveUploadedFile(context.Background(), UploadedFile{Filename: "safe.txt", Content: []byte("x")})
	if !errors.Is(err, storageErr) {
		t.Fatalf("storage error = %v, want %v", err, storageErr)
	}
}

type failingUploadStorage struct {
	err error
}

func (s failingUploadStorage) SaveUpload(context.Context, UploadedFile) (StoredUpload, error) {
	return StoredUpload{}, s.err
}
