package fields

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileConfig configures file fields.
type FileConfig struct {
	UploadTo string
	Generate func(filename string) (string, error)
}

// FileField stores a file path.
type FileField struct {
	*BaseField
	config FileConfig
}

// NewFileField creates a file field.
func NewFileField(options Options, config FileConfig) *FileField {
	return &FileField{BaseField: NewBaseField("file", options, map[string]string{"postgres": "varchar(255)", "sqlite": "text"}), config: config}
}

func (f *FileField) GeneratePath(filename string) (string, error) {
	target := ""
	if f.config.Generate != nil {
		generated, err := f.config.Generate(filename)
		if err != nil {
			return "", err
		}
		target = generated
	} else {
		if !safeRelativePath(filename) {
			return "", fmt.Errorf("%w: unsafe file path %q", ErrInvalidField, filename)
		}
		target = filepath.Join(f.config.UploadTo, filename)
	}
	if !safeRelativePath(target) {
		return "", fmt.Errorf("%w: unsafe file path %q", ErrInvalidField, target)
	}
	return target, nil
}

func (f *FileField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || f.emptyAllowed(value) {
		return err
	}
	path, ok := value.(string)
	if !ok || !safeRelativePath(path) {
		return fmt.Errorf("%w: %s must be safe relative path", ErrValidation, f.Name())
	}
	return nil
}

func (f *FileField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	return value.(string), nil
}

func (f *FileField) FromDB(value any) (any, error) {
	return f.ToDB(value)
}

func (f *FileField) Clone() Field {
	return &FileField{BaseField: f.BaseField.Clone().(*BaseField), config: f.config}
}

func (f *FileField) emptyAllowed(value any) bool {
	return value == nil && f.options.Null || value == "" && f.options.Blank
}

// ImageMetadata contains inspected image details.
type ImageMetadata struct {
	Width  int
	Height int
	Format string
}

// ImageField stores image paths and exposes an inspector hook.
type ImageField struct {
	*FileField
	inspect func(string) (ImageMetadata, error)
}

// NewImageField creates an image field.
func NewImageField(options Options, config FileConfig, inspector func(string) (ImageMetadata, error)) *ImageField {
	return &ImageField{FileField: NewFileField(options, config), inspect: inspector}
}

// Inspect returns image metadata through the configured hook.
func (f *ImageField) Inspect(path string) (ImageMetadata, error) {
	if f.inspect == nil {
		return ImageMetadata{}, fmt.Errorf("%w: image inspector is not configured", ErrInvalidField)
	}
	return f.inspect(path)
}

func (f *ImageField) Clone() Field {
	return &ImageField{FileField: f.FileField.Clone().(*FileField), inspect: f.inspect}
}

// FilePathConfig configures a FilePathField.
type FilePathConfig struct {
	Root string
}

// FilePathField validates paths under a root directory.
type FilePathField struct {
	*BaseField
	config FilePathConfig
}

// NewFilePathField creates a file path field.
func NewFilePathField(options Options, config FilePathConfig) *FilePathField {
	return &FilePathField{BaseField: NewBaseField("filepath", options, map[string]string{"postgres": "varchar(255)", "sqlite": "text"}), config: config}
}

func (f *FilePathField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || value == nil && f.options.Null {
		return err
	}
	path, ok := value.(string)
	if !ok || !f.insideRoot(path) {
		return fmt.Errorf("%w: %s must be inside root", ErrValidation, f.Name())
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return nil
}

func (f *FilePathField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	return value.(string), nil
}

func (f *FilePathField) FromDB(value any) (any, error) {
	return f.ToDB(value)
}

func (f *FilePathField) Clone() Field {
	return &FilePathField{BaseField: f.BaseField.Clone().(*BaseField), config: f.config}
}

func (f *FilePathField) insideRoot(path string) bool {
	root, err := filepath.Abs(f.config.Root)
	if err != nil {
		return false
	}
	target, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return target == root || strings.HasPrefix(target, root+string(filepath.Separator))
}

func safeRelativePath(path string) bool {
	if path == "" || filepath.IsAbs(path) {
		return false
	}
	cleaned := filepath.Clean(path)
	if cleaned == "." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return false
	}
	for _, segment := range strings.Split(cleaned, string(filepath.Separator)) {
		if segment == ".." {
			return false
		}
	}
	return true
}
