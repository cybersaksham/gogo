package api

import (
	"context"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var routeParamPattern = regexp.MustCompile(`<([A-Za-z_][A-Za-z0-9_]*):([A-Za-z_][A-Za-z0-9_]*)>`)

// OpenAPIOptions configures OpenAPI generation.
type OpenAPIOptions struct {
	Title          string
	Version        string
	Router         *Router
	Serializer     *Serializer
	Pagination     any
	Authentication []string
	AdminDocsURL   string
}

// OpenAPISpec is an OpenAPI 3.1 document.
type OpenAPISpec struct {
	OpenAPI      string                     `json:"openapi"`
	Info         OpenAPIInfo                `json:"info"`
	ExternalDocs OpenAPIExternalDocs        `json:"externalDocs,omitempty"`
	Paths        map[string]OpenAPIPathItem `json:"paths"`
	Components   OpenAPIComponents          `json:"components"`
	Security     []map[string][]string      `json:"security,omitempty"`
}

// OpenAPIInfo stores API title and version.
type OpenAPIInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// OpenAPIExternalDocs stores optional documentation links.
type OpenAPIExternalDocs struct {
	URL string `json:"url,omitempty"`
}

// OpenAPIPathItem maps HTTP methods to operations.
type OpenAPIPathItem map[string]OpenAPIOperation

// OpenAPIOperation describes one operation.
type OpenAPIOperation struct {
	OperationID string                `json:"operationId"`
	Parameters  []OpenAPIParameter    `json:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody   `json:"requestBody,omitempty"`
	Responses   map[string]any        `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
}

// OperationMetadata documents a custom or raw API route for OpenAPI generation.
type OperationMetadata struct {
	Summary     string
	Description string
	Tags        []string
	Responses   map[int]ResponseSchema
}

// ResponseSchema documents one OpenAPI response for a custom or raw API route.
type ResponseSchema struct {
	Description string
	ContentType string
	Schema      map[string]any
}

// OpenAPIParameter describes an operation parameter.
type OpenAPIParameter struct {
	In       string         `json:"in"`
	Name     string         `json:"name"`
	Required bool           `json:"required"`
	Schema   map[string]any `json:"schema"`
}

// OpenAPIRequestBody describes operation request content.
type OpenAPIRequestBody struct {
	Required bool                        `json:"required"`
	Content  map[string]OpenAPIMediaType `json:"content"`
}

// OpenAPIMediaType wraps a schema.
type OpenAPIMediaType struct {
	Schema map[string]any `json:"schema"`
}

// OpenAPIComponents stores reusable schemas, responses, and security schemes.
type OpenAPIComponents struct {
	Schemas         map[string]any            `json:"schemas"`
	Responses       map[string]any            `json:"responses"`
	SecuritySchemes map[string]map[string]any `json:"securitySchemes,omitempty"`
}

// GenerateOpenAPI builds an OpenAPI 3.1 document.
func GenerateOpenAPI(options OpenAPIOptions) OpenAPISpec {
	title := stringDefault(options.Title, "API")
	version := stringDefault(options.Version, "1.0.0")
	components := OpenAPIComponents{
		Schemas: map[string]any{
			"Default":      serializerOpenAPISchema(options.Serializer),
			"Paginated":    paginatedOpenAPISchema(),
			"UploadedFile": uploadedFileOpenAPISchema(),
			"Error":        errorOpenAPISchema(),
		},
		Responses: map[string]any{
			"Error": map[string]any{
				"description": "Error",
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/Error"},
					},
				},
			},
		},
		SecuritySchemes: securitySchemes(options.Authentication),
	}
	spec := OpenAPISpec{
		OpenAPI:    "3.1.0",
		Info:       OpenAPIInfo{Title: title, Version: version},
		Paths:      map[string]OpenAPIPathItem{},
		Components: components,
		Security:   securityRequirements(options.Authentication),
	}
	if options.AdminDocsURL != "" {
		spec.ExternalDocs = OpenAPIExternalDocs{URL: options.AdminDocsURL}
	}
	if options.Router != nil {
		for _, route := range options.Router.Routes() {
			path, params := openAPIPath(route.Pattern)
			if spec.Paths[path] == nil {
				spec.Paths[path] = OpenAPIPathItem{}
			}
			for _, method := range route.Methods {
				spec.Paths[path][strings.ToLower(method)] = openAPIOperation(route, method, params, options)
			}
		}
	}
	return spec
}

// OpenAPIJSONView returns an API view that serves an OpenAPI document.
func OpenAPIJSONView(spec OpenAPISpec) View {
	return func(context.Context, *Request) Response {
		return JSON(http.StatusOK, spec)
	}
}

func openAPIOperation(route Route, method string, params []string, options OpenAPIOptions) OpenAPIOperation {
	operation := OpenAPIOperation{
		OperationID: route.Name,
		Parameters:  openAPIParameters(params),
		Responses:   openAPIResponses(route, method, options),
		Security:    securityRequirements(options.Authentication),
		Summary:     route.Action,
		Tags:        []string{routeTag(route.Name)},
	}
	if requestBodyRequired(method) {
		operation.RequestBody = &OpenAPIRequestBody{
			Required: true,
			Content: map[string]OpenAPIMediaType{
				"application/json": {Schema: map[string]any{"$ref": "#/components/schemas/Default"}},
			},
		}
	}
	applyOperationMetadata(&operation, route.Metadata)
	return operation
}

func openAPIResponses(route Route, method string, options OpenAPIOptions) map[string]any {
	responses := map[string]any{
		"400":     map[string]any{"$ref": "#/components/responses/Error"},
		"401":     map[string]any{"$ref": "#/components/responses/Error"},
		"403":     map[string]any{"$ref": "#/components/responses/Error"},
		"404":     map[string]any{"$ref": "#/components/responses/Error"},
		"429":     map[string]any{"$ref": "#/components/responses/Error"},
		"default": map[string]any{"$ref": "#/components/responses/Error"},
	}
	status := "200"
	description := "OK"
	schema := map[string]any{"$ref": "#/components/schemas/Default"}
	if route.Action == "list" && paginationMetadataName(options.Pagination) != "" {
		schema = map[string]any{"$ref": "#/components/schemas/Paginated"}
	}
	switch method {
	case http.MethodPost:
		status = "201"
		description = "Created"
	case http.MethodDelete:
		responses["204"] = map[string]any{"description": "No Content"}
		return responses
	}
	responses[status] = map[string]any{
		"content": map[string]any{
			"application/json": map[string]any{"schema": schema},
		},
		"description": description,
	}
	return responses
}

func applyOperationMetadata(operation *OpenAPIOperation, metadata OperationMetadata) {
	if metadata.Summary != "" {
		operation.Summary = metadata.Summary
	}
	if metadata.Description != "" {
		operation.Description = metadata.Description
	}
	if len(metadata.Tags) > 0 {
		operation.Tags = append([]string(nil), metadata.Tags...)
	}
	if len(metadata.Responses) == 0 {
		return
	}
	for _, status := range sortedResponseStatuses(metadata.Responses) {
		operation.Responses[strconv.Itoa(status)] = openAPIResponseFromSchema(metadata.Responses[status])
	}
}

func openAPIResponseFromSchema(response ResponseSchema) map[string]any {
	description := response.Description
	if description == "" {
		description = "Response"
	}
	documented := map[string]any{"description": description}
	if len(response.Schema) == 0 {
		return documented
	}
	contentType := response.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	documented["content"] = map[string]any{
		contentType: map[string]any{"schema": cloneAnyMap(response.Schema)},
	}
	return documented
}

func sortedResponseStatuses(responses map[int]ResponseSchema) []int {
	statuses := make([]int, 0, len(responses))
	for status := range responses {
		statuses = append(statuses, status)
	}
	sort.Ints(statuses)
	return statuses
}

func cloneOperationMetadata(metadata OperationMetadata) OperationMetadata {
	copied := OperationMetadata{
		Summary:     metadata.Summary,
		Description: metadata.Description,
		Tags:        append([]string(nil), metadata.Tags...),
	}
	if len(metadata.Responses) > 0 {
		copied.Responses = make(map[int]ResponseSchema, len(metadata.Responses))
		for status, response := range metadata.Responses {
			if len(response.Schema) > 0 {
				response.Schema = cloneAnyMap(response.Schema)
			}
			copied.Responses[status] = response
		}
	}
	return copied
}

func openAPIPath(pattern string) (string, []string) {
	params := make([]string, 0)
	path := routeParamPattern.ReplaceAllStringFunc(pattern, func(value string) string {
		matches := routeParamPattern.FindStringSubmatch(value)
		if len(matches) != 3 {
			return value
		}
		params = append(params, matches[2])
		return "{" + matches[2] + "}"
	})
	return path, params
}

func openAPIParameters(params []string) []OpenAPIParameter {
	parameters := make([]OpenAPIParameter, 0, len(params))
	for _, param := range params {
		parameters = append(parameters, OpenAPIParameter{
			In:       "path",
			Name:     param,
			Required: true,
			Schema:   map[string]any{"type": "string"},
		})
	}
	return parameters
}

func serializerOpenAPISchema(serializer *Serializer) map[string]any {
	properties := map[string]any{}
	var required []string
	if serializer != nil {
		for _, field := range serializer.fields {
			properties[field.Name] = fieldOpenAPISchema(field)
			if field.Options.Required && !field.Options.ReadOnly {
				required = append(required, field.Name)
			}
		}
	}
	schema := map[string]any{"type": "object", "properties": properties}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func fieldOpenAPISchema(field SerializerField) map[string]any {
	schema := map[string]any{}
	switch field.Kind {
	case "boolean":
		schema["type"] = "boolean"
	case "integer", "primary_key_related":
		schema["type"] = "integer"
		schema["format"] = "int64"
	case "float":
		schema["type"] = "number"
		schema["format"] = "double"
	case "list", "multiple_choice":
		schema["type"] = "array"
		schema["items"] = map[string]any{"type": "string"}
	case "dict", "json", "nested":
		schema["type"] = "object"
	default:
		schema["type"] = "string"
	}
	if len(field.Choices) > 0 {
		schema["enum"] = append([]string(nil), field.Choices...)
	}
	if field.Options.ReadOnly {
		schema["readOnly"] = true
	}
	if field.Options.WriteOnly {
		schema["writeOnly"] = true
	}
	return schema
}

func paginatedOpenAPISchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"count":    map[string]any{"type": "integer"},
			"next":     map[string]any{"type": "string"},
			"previous": map[string]any{"type": "string"},
			"results":  map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/Default"}},
		},
	}
}

func uploadedFileOpenAPISchema() map[string]any {
	return map[string]any{"type": "string", "format": "binary"}
}

func errorOpenAPISchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"error": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code":       map[string]any{"type": "string"},
					"message":    map[string]any{"type": "string"},
					"fields":     map[string]any{"type": "object"},
					"request_id": map[string]any{"type": "string"},
				},
			},
		},
	}
}

func securitySchemes(values []string) map[string]map[string]any {
	schemes := map[string]map[string]any{}
	for _, value := range values {
		switch strings.ToLower(value) {
		case "token":
			schemes["TokenAuth"] = map[string]any{"type": "apiKey", "in": "header", "name": "Authorization"}
		case "session":
			schemes["SessionAuth"] = map[string]any{"type": "apiKey", "in": "cookie", "name": "sessionid"}
		}
	}
	if len(schemes) == 0 {
		return nil
	}
	return schemes
}

func securityRequirements(values []string) []map[string][]string {
	var requirements []map[string][]string
	for _, value := range values {
		switch strings.ToLower(value) {
		case "token":
			requirements = append(requirements, map[string][]string{"TokenAuth": {}})
		case "session":
			requirements = append(requirements, map[string][]string{"SessionAuth": {}})
		}
	}
	return requirements
}

func routeTag(name string) string {
	if tag, _, ok := strings.Cut(name, "-"); ok && tag != "" {
		return tag
	}
	return "default"
}

func requestBodyRequired(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}
