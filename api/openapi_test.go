package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestOpenAPIGenerationCoversViewSetAuthAndErrors(t *testing.T) {
	router := NewRouter()
	viewset := &ModelViewSet{Store: newMemoryViewSetStore()}
	viewset.RegisterAction("publish", ViewSetAction{Handler: func(context.Context, *Request) Response { return NoContent() }, Detail: true})
	if err := router.Register("posts", "post", viewset); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	serializer := NewSerializer(
		IntegerField("id", FieldOptions{ReadOnly: true}),
		StringField("title", FieldOptions{Required: true}),
	)

	spec := GenerateOpenAPI(OpenAPIOptions{
		Title:          "Gogo API",
		Version:        "v1",
		Router:         router,
		Serializer:     serializer,
		Pagination:     PageNumberPagination{PageSize: 20},
		Authentication: []string{"token"},
		AdminDocsURL:   "/admin/docs/openapi/",
	})

	if spec.OpenAPI != "3.1.0" || spec.Info.Title != "Gogo API" || spec.ExternalDocs.URL != "/admin/docs/openapi/" {
		t.Fatalf("spec header = %#v", spec)
	}
	if spec.Paths["/posts/"]["get"].OperationID != "post-list" || spec.Paths["/posts/"]["post"].RequestBody == nil {
		t.Fatalf("collection path = %#v", spec.Paths["/posts/"])
	}
	if spec.Paths["/posts/{id}/"]["get"].Parameters[0].Name != "id" || spec.Paths["/posts/{id}/"]["put"].RequestBody == nil {
		t.Fatalf("detail path = %#v", spec.Paths["/posts/{id}/"])
	}
	if spec.Components.SecuritySchemes["TokenAuth"]["type"] != "apiKey" || len(spec.Security) != 1 {
		t.Fatalf("security = %#v %#v", spec.Components.SecuritySchemes, spec.Security)
	}
	if spec.Components.Schemas["Error"].(map[string]any)["type"] != "object" {
		t.Fatalf("error schema = %#v", spec.Components.Schemas["Error"])
	}

	custom, err := json.MarshalIndent(spec.Paths["/posts/{id}/publish/"], "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	golden := `{
  "get": {
    "operationId": "post-publish",
    "parameters": [
      {
        "in": "path",
        "name": "id",
        "required": true,
        "schema": {
          "type": "string"
        }
      }
    ],
    "responses": {
      "200": {
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/Default"
            }
          }
        },
        "description": "OK"
      },
      "400": {
        "$ref": "#/components/responses/Error"
      },
      "401": {
        "$ref": "#/components/responses/Error"
      },
      "403": {
        "$ref": "#/components/responses/Error"
      },
      "404": {
        "$ref": "#/components/responses/Error"
      },
      "429": {
        "$ref": "#/components/responses/Error"
      },
      "default": {
        "$ref": "#/components/responses/Error"
      }
    },
    "security": [
      {
        "TokenAuth": []
      }
    ],
    "summary": "publish",
    "tags": [
      "post"
    ]
  }
}`
	if strings.TrimSpace(string(custom)) != golden {
		t.Fatalf("custom path golden mismatch:\n%s", custom)
	}
}

func TestOpenAPIIncludesDocumentedRawHandlers(t *testing.T) {
	router := NewRouter()
	if err := router.HandleHTTP("legacy-report", "reports/<str:id>", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), OperationMetadata{
		Summary:     "Legacy report",
		Description: "Existing report endpoint preserved during migration.",
		Tags:        []string{"legacy"},
		Responses: map[int]ResponseSchema{
			http.StatusOK: {
				Description: "Legacy payload",
				ContentType: "application/json",
				Schema:      map[string]any{"type": "object"},
			},
			http.StatusNoContent: {Description: "No content"},
		},
	}, http.MethodGet); err != nil {
		t.Fatalf("HandleHTTP() error = %v", err)
	}

	spec := GenerateOpenAPI(OpenAPIOptions{Router: router})
	operation := spec.Paths["/reports/{id}/"]["get"]
	if operation.OperationID != "legacy-report" || operation.Summary != "Legacy report" || operation.Description != "Existing report endpoint preserved during migration." {
		t.Fatalf("operation header = %#v", operation)
	}
	if len(operation.Tags) != 1 || operation.Tags[0] != "legacy" {
		t.Fatalf("tags = %#v", operation.Tags)
	}
	if operation.Parameters[0].Name != "id" {
		t.Fatalf("parameters = %#v", operation.Parameters)
	}
	okResponse := operation.Responses["200"].(map[string]any)
	if okResponse["description"] != "Legacy payload" {
		t.Fatalf("200 response = %#v", okResponse)
	}
	content := okResponse["content"].(map[string]any)["application/json"].(map[string]any)
	if content["schema"].(map[string]any)["type"] != "object" {
		t.Fatalf("200 content = %#v", content)
	}
	if operation.Responses["204"].(map[string]any)["description"] != "No content" {
		t.Fatalf("204 response = %#v", operation.Responses["204"])
	}
}

func TestOpenAPIJSONViewReturnsSpec(t *testing.T) {
	spec := GenerateOpenAPI(OpenAPIOptions{Title: "Gogo API", Version: "v1"})
	response := OpenAPIJSONView(spec)(context.Background(), nil)
	if response.status != http.StatusOK || response.body.(OpenAPISpec).Info.Title != "Gogo API" {
		t.Fatalf("response = %#v", response)
	}
}
