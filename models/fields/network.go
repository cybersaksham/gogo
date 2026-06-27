package fields

import (
	"fmt"
	"net"
)

type IPProtocol string

const (
	BothIP IPProtocol = "both"
	IPv4   IPProtocol = "ipv4"
	IPv6   IPProtocol = "ipv6"
)

// IPAddressConfig configures generic IP address fields.
type IPAddressConfig struct {
	Protocol         IPProtocol
	UnpackIPv4Mapped bool
}

// GenericIPAddressField stores IPv4 and IPv6 addresses.
type GenericIPAddressField struct {
	*BaseField
	config IPAddressConfig
}

// NewGenericIPAddressField creates an IP address field.
func NewGenericIPAddressField(options Options, config IPAddressConfig) *GenericIPAddressField {
	if config.Protocol == "" {
		config.Protocol = BothIP
	}
	return &GenericIPAddressField{BaseField: NewBaseField("ip_address", options, map[string]string{"postgres": "inet", "sqlite": "text"}), config: config}
}

func (f *GenericIPAddressField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || f.emptyAllowed(value) {
		return err
	}
	if _, err := f.normalized(value); err != nil {
		return err
	}
	return nil
}

func (f *GenericIPAddressField) ToDB(value any) (any, error) {
	return f.normalized(value)
}

func (f *GenericIPAddressField) FromDB(value any) (any, error) {
	return f.normalized(value)
}

func (f *GenericIPAddressField) Clone() Field {
	return &GenericIPAddressField{BaseField: f.BaseField.Clone().(*BaseField), config: f.config}
}

func (f *GenericIPAddressField) normalized(value any) (string, error) {
	raw, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s must be IP string", ErrValidation, f.Name())
	}
	ip := net.ParseIP(raw)
	if ip == nil {
		return "", fmt.Errorf("%w: %s must be IP address", ErrValidation, f.Name())
	}
	is4 := ip.To4() != nil
	switch f.config.Protocol {
	case IPv4:
		if !is4 {
			return "", fmt.Errorf("%w: %s must be IPv4", ErrValidation, f.Name())
		}
	case IPv6:
		if is4 {
			return "", fmt.Errorf("%w: %s must be IPv6", ErrValidation, f.Name())
		}
	}
	if f.config.UnpackIPv4Mapped && is4 {
		return ip.To4().String(), nil
	}
	return ip.String(), nil
}

func (f *GenericIPAddressField) emptyAllowed(value any) bool {
	return value == nil && f.options.Null || value == "" && f.options.Blank
}
