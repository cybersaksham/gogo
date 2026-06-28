package api

import (
	"context"
	"fmt"
	"mime"
	"net"
	"strings"
)

// VersioningConfig configures allowed and default API versions.
type VersioningConfig struct {
	AllowedVersions []string
	DefaultVersion  string
}

// VersioningStrategy resolves an API version for one request.
type VersioningStrategy interface {
	ResolveVersion(*Request) (string, error)
}

// URLPathVersioning resolves versions from the first URL path segment.
type URLPathVersioning struct {
	Config VersioningConfig
}

// NamespaceVersioning resolves versions from a path parameter.
type NamespaceVersioning struct {
	Config VersioningConfig
	Param  string
}

// HostNameVersioning resolves versions from the first hostname label.
type HostNameVersioning struct {
	Config VersioningConfig
}

// QueryParameterVersioning resolves versions from a query parameter.
type QueryParameterVersioning struct {
	Config VersioningConfig
	Param  string
}

// AcceptHeaderVersioning resolves versions from an Accept header version parameter.
type AcceptHeaderVersioning struct {
	Config VersioningConfig
}

// ResolveVersion resolves URL path versions.
func (v URLPathVersioning) ResolveVersion(request *Request) (string, error) {
	path := strings.Trim(request.Raw().URL.Path, "/")
	if path == "" {
		return resolveConfiguredVersion("", v.Config)
	}
	segment := strings.Split(path, "/")[0]
	if strings.HasPrefix(strings.ToLower(segment), "v") {
		return resolveConfiguredVersion(segment, v.Config)
	}
	return resolveConfiguredVersion("", v.Config)
}

// ResolveVersion resolves namespace path parameter versions.
func (v NamespaceVersioning) ResolveVersion(request *Request) (string, error) {
	param := stringDefault(v.Param, "version")
	return resolveConfiguredVersion(request.PathParam(param), v.Config)
}

// ResolveVersion resolves hostname versions.
func (v HostNameVersioning) ResolveVersion(request *Request) (string, error) {
	host := request.Raw().Host
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	labels := strings.Split(host, ".")
	if len(labels) > 0 && strings.HasPrefix(strings.ToLower(labels[0]), "v") {
		return resolveConfiguredVersion(labels[0], v.Config)
	}
	return resolveConfiguredVersion("", v.Config)
}

// ResolveVersion resolves query parameter versions.
func (v QueryParameterVersioning) ResolveVersion(request *Request) (string, error) {
	param := stringDefault(v.Param, "version")
	return resolveConfiguredVersion(request.QueryParam(param), v.Config)
}

// ResolveVersion resolves Accept header versions.
func (v AcceptHeaderVersioning) ResolveVersion(request *Request) (string, error) {
	header := request.Raw().Header.Get("Accept")
	for _, value := range strings.Split(header, ",") {
		mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(value))
		if err != nil || mediaType == "" {
			continue
		}
		if version := params["version"]; version != "" {
			return resolveConfiguredVersion(version, v.Config)
		}
		if vendorVersion := vendorMediaVersion(mediaType); vendorVersion != "" {
			return resolveConfiguredVersion(vendorVersion, v.Config)
		}
	}
	return resolveConfiguredVersion("", v.Config)
}

// VersionRequest creates an APIView lifecycle hook that attaches the resolved version.
func VersionRequest(strategy VersioningStrategy) RequestHook {
	return func(_ context.Context, request *Request) error {
		if strategy == nil {
			return nil
		}
		version, err := strategy.ResolveVersion(request)
		if err != nil {
			return err
		}
		request.WithVersion(version)
		return nil
	}
}

func resolveConfiguredVersion(candidate string, config VersioningConfig) (string, error) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		candidate = config.DefaultVersion
	}
	if candidate == "" {
		return "", nil
	}
	if len(config.AllowedVersions) == 0 {
		return candidate, nil
	}
	for _, allowed := range config.AllowedVersions {
		if candidate == allowed {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%w: version %q is not allowed", ErrVersion, candidate)
}

func vendorMediaVersion(mediaType string) string {
	parts := strings.Split(mediaType, ".")
	last := parts[len(parts)-1]
	if index := strings.Index(last, "+"); index >= 0 {
		last = last[:index]
	}
	if strings.HasPrefix(strings.ToLower(last), "v") {
		return last
	}
	return ""
}
