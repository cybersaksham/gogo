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

func TestOpenAPIJSONViewReturnsSpec(t *testing.T) {
	spec := GenerateOpenAPI(OpenAPIOptions{Title: "Gogo API", Version: "v1"})
	response := OpenAPIJSONView(spec)(context.Background(), nil)
	if response.status != http.StatusOK || response.body.(OpenAPISpec).Info.Title != "Gogo API" {
		t.Fatalf("response = %#v", response)
	}
}
