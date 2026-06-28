package files

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

// MemoryUploadHandler keeps uploads in memory.
type MemoryUploadHandler struct {
	MaxMemory int64
	MaxSize   int64
}

func (h MemoryUploadHandler) Handle(ctx context.Context, source UploadSource) (*Upload, error) {
	name, err := NormalizeName(source.Name)
	if err != nil {
		return nil, err
	}
	content, err := readAllLimited(ctx, source.Reader, h.limit())
	if err != nil {
		return nil, err
	}
	if h.MaxMemory > 0 && int64(len(content)) > h.MaxMemory {
		return nil, fmt.Errorf("%w: memory limit exceeded", ErrUploadTooLarge)
	}
	return &Upload{Name: name, ContentType: source.ContentType, Size: int64(len(content)), Content: content}, nil
}

func (h MemoryUploadHandler) limit() int64 {
	if h.MaxSize > 0 {
		return h.MaxSize
	}
	return h.MaxMemory
}

// TemporaryUploadHandler writes uploads to temporary files.
type TemporaryUploadHandler struct {
	Dir     string
	MaxSize int64
}

func (h TemporaryUploadHandler) Handle(ctx context.Context, source UploadSource) (*Upload, error) {
	name, err := NormalizeName(source.Name)
	if err != nil {
		return nil, err
	}
	dir := h.Dir
	if dir == "" {
		dir = os.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	pattern := "upload-*-" + path.Base(name)
	file, err := os.CreateTemp(dir, filepath.Base(pattern))
	if err != nil {
		return nil, err
	}
	tempPath := file.Name()
	size, copyErr := copyLimited(ctx, file, source.Reader, h.MaxSize)
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil {
		os.Remove(tempPath)
		if copyErr != nil {
			return nil, copyErr
		}
		return nil, closeErr
	}
	return &Upload{Name: name, ContentType: source.ContentType, Size: size, TemporaryPath: tempPath}, nil
}

// ChunkedUploadHandler streams uploads to storage in chunks.
type ChunkedUploadHandler struct {
	Storage   Storage
	ChunkSize int
	MaxSize   int64
}

func (h ChunkedUploadHandler) Handle(ctx context.Context, source UploadSource) (*Upload, error) {
	if h.Storage == nil {
		return nil, fmt.Errorf("%w: storage is required", ErrInvalidUpload)
	}
	name, err := NormalizeName(source.Name)
	if err != nil {
		return nil, err
	}
	content, err := readAllLimitedWithChunk(ctx, source.Reader, h.MaxSize, h.ChunkSize)
	if err != nil {
		return nil, err
	}
	storedName, err := h.Storage.Save(ctx, name, bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	return &Upload{Name: name, ContentType: source.ContentType, Size: int64(len(content)), StoredName: storedName}, nil
}

func readAllLimited(ctx context.Context, reader io.Reader, maxSize int64) ([]byte, error) {
	return readAllLimitedWithChunk(ctx, reader, maxSize, 32*1024)
}

func readAllLimitedWithChunk(ctx context.Context, reader io.Reader, maxSize int64, chunkSize int) ([]byte, error) {
	if reader == nil {
		reader = bytes.NewReader(nil)
	}
	if chunkSize <= 0 {
		chunkSize = 32 * 1024
	}
	var buffer bytes.Buffer
	chunk := make([]byte, chunkSize)
	for {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrUploadInterrupted, err)
		}
		n, err := reader.Read(chunk)
		if n > 0 {
			if maxSize > 0 && int64(buffer.Len()+n) > maxSize {
				return nil, fmt.Errorf("%w: limit %d", ErrUploadTooLarge, maxSize)
			}
			buffer.Write(chunk[:n])
		}
		if err == io.EOF {
			return buffer.Bytes(), nil
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrUploadInterrupted, err)
		}
	}
}

func copyLimited(ctx context.Context, writer io.Writer, reader io.Reader, maxSize int64) (int64, error) {
	content, err := readAllLimited(ctx, reader, maxSize)
	if err != nil {
		return 0, err
	}
	written, err := writer.Write(content)
	if err != nil {
		return 0, err
	}
	return int64(written), nil
}
