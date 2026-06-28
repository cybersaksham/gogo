package files

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestUploadHandlersMemoryTemporaryChunkedAndLimits(t *testing.T) {
	ctx := context.Background()

	memory := MemoryUploadHandler{MaxMemory: 32, MaxSize: 32}
	upload, err := memory.Handle(ctx, UploadSource{
		Name:        "avatar.png",
		ContentType: "image/png",
		Reader:      strings.NewReader("png"),
	})
	if err != nil {
		t.Fatalf("memory Handle() error = %v", err)
	}
	if upload.Name != "avatar.png" || upload.Size != 3 || string(upload.Content) != "png" {
		t.Fatalf("memory upload = %#v", upload)
	}
	if _, err := memory.Handle(ctx, UploadSource{Name: "big.txt", Reader: strings.NewReader(strings.Repeat("x", 33))}); !errors.Is(err, ErrUploadTooLarge) {
		t.Fatalf("memory limit error = %v, want ErrUploadTooLarge", err)
	}

	tempDir := t.TempDir()
	temporary := TemporaryUploadHandler{Dir: tempDir, MaxSize: 64}
	tempUpload, err := temporary.Handle(ctx, UploadSource{Name: "docs/report.txt", ContentType: "text/plain", Reader: strings.NewReader("temporary")})
	if err != nil {
		t.Fatalf("temporary Handle() error = %v", err)
	}
	if tempUpload.TemporaryPath == "" || tempUpload.Size != int64(len("temporary")) {
		t.Fatalf("temporary upload = %#v", tempUpload)
	}
	if _, err := os.Stat(tempUpload.TemporaryPath); err != nil {
		t.Fatalf("temporary file missing: %v", err)
	}

	storage := NewLocalStorage(t.TempDir(), LocalOptions{})
	chunked := ChunkedUploadHandler{Storage: storage, ChunkSize: 2, MaxSize: 64}
	chunkUpload, err := chunked.Handle(ctx, UploadSource{Name: "chunks/data.txt", ContentType: "text/plain", Reader: strings.NewReader("abcdef")})
	if err != nil {
		t.Fatalf("chunked Handle() error = %v", err)
	}
	reader, err := storage.Open(ctx, chunkUpload.StoredName)
	if err != nil {
		t.Fatalf("Open(chunked stored) error = %v", err)
	}
	content, _ := io.ReadAll(reader)
	reader.Close()
	if string(content) != "abcdef" {
		t.Fatalf("chunked content = %q", content)
	}
}

func TestUploadInterruptionAndValidators(t *testing.T) {
	ctx := context.Background()
	handler := TemporaryUploadHandler{Dir: t.TempDir(), MaxSize: 64}
	if _, err := handler.Handle(ctx, UploadSource{Name: "broken.txt", Reader: interruptedReader{}}); !errors.Is(err, ErrUploadInterrupted) {
		t.Fatalf("interrupted upload error = %v, want ErrUploadInterrupted", err)
	}

	upload := &Upload{Name: "avatar.png", ContentType: "image/png", Size: 10, Content: []byte("png")}
	err := ValidateUpload(upload,
		MaxSizeValidator(20),
		ContentTypeValidator("image/png"),
		ExtensionValidator(".png", ".jpg"),
		ImageDimensionsValidator(func(*Upload) (int, int, error) { return 100, 80, nil }, 200, 200),
	)
	if err != nil {
		t.Fatalf("ValidateUpload() error = %v", err)
	}
	if err := ValidateUpload(upload, MaxSizeValidator(5)); !errors.Is(err, ErrUploadTooLarge) {
		t.Fatalf("MaxSizeValidator error = %v, want ErrUploadTooLarge", err)
	}
	if err := ValidateUpload(upload, ContentTypeValidator("image/jpeg")); !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("ContentTypeValidator error = %v, want ErrInvalidUpload", err)
	}
	if err := ValidateUpload(upload, ExtensionValidator(".gif")); !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("ExtensionValidator error = %v, want ErrInvalidUpload", err)
	}
	if err := ValidateUpload(upload, ImageDimensionsValidator(func(*Upload) (int, int, error) { return 500, 80, nil }, 200, 200)); !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("ImageDimensionsValidator error = %v, want ErrInvalidUpload", err)
	}
}

type interruptedReader struct{}

func (interruptedReader) Read([]byte) (int, error) {
	return 0, errors.New("client disconnected")
}
