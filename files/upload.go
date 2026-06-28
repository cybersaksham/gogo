package files

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// UploadSource is an incoming upload stream.
type UploadSource struct {
	Name        string
	ContentType string
	Reader      io.Reader
	Size        int64
}

// Upload is a handled upload result.
type Upload struct {
	Name          string
	ContentType   string
	Size          int64
	Content       []byte
	TemporaryPath string
	StoredName    string
}

func (u *Upload) Open() (io.ReadCloser, error) {
	if u == nil {
		return nil, fmt.Errorf("%w: upload is nil", ErrInvalidUpload)
	}
	if u.TemporaryPath != "" {
		return os.Open(u.TemporaryPath)
	}
	return io.NopCloser(bytes.NewReader(u.Content)), nil
}

// UploadValidator validates a handled upload.
type UploadValidator func(*Upload) error

func ValidateUpload(upload *Upload, validators ...UploadValidator) error {
	for _, validator := range validators {
		if validator == nil {
			continue
		}
		if err := validator(upload); err != nil {
			return err
		}
	}
	return nil
}

func MaxSizeValidator(maxSize int64) UploadValidator {
	return func(upload *Upload) error {
		if upload != nil && maxSize > 0 && upload.Size > maxSize {
			return fmt.Errorf("%w: %d > %d", ErrUploadTooLarge, upload.Size, maxSize)
		}
		return nil
	}
}

func ContentTypeValidator(allowed ...string) UploadValidator {
	allowedSet := stringSet(allowed)
	return func(upload *Upload) error {
		if upload == nil || !allowedSet[upload.ContentType] {
			return fmt.Errorf("%w: content type %q", ErrInvalidUpload, uploadContentType(upload))
		}
		return nil
	}
}

func ExtensionValidator(allowed ...string) UploadValidator {
	allowedSet := map[string]bool{}
	for _, extension := range allowed {
		if extension == "" {
			continue
		}
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
		allowedSet[strings.ToLower(extension)] = true
	}
	return func(upload *Upload) error {
		extension := strings.ToLower(path.Ext(uploadName(upload)))
		if !allowedSet[extension] {
			return fmt.Errorf("%w: extension %q", ErrInvalidUpload, extension)
		}
		return nil
	}
}

func ImageDimensionsValidator(inspect func(*Upload) (int, int, error), maxWidth, maxHeight int) UploadValidator {
	return func(upload *Upload) error {
		if inspect == nil {
			return nil
		}
		width, height, err := inspect(upload)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUpload, err)
		}
		if maxWidth > 0 && width > maxWidth || maxHeight > 0 && height > maxHeight {
			return fmt.Errorf("%w: image dimensions %dx%d", ErrInvalidUpload, width, height)
		}
		return nil
	}
}

func uploadName(upload *Upload) string {
	if upload == nil {
		return ""
	}
	return upload.Name
}

func uploadContentType(upload *Upload) string {
	if upload == nil {
		return ""
	}
	return upload.ContentType
}

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
