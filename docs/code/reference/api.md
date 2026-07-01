# API Reference

The API package provides request/response wrappers, serializers, parser and renderer registries, class-based views, ViewSets, routers, authentication, permissions, throttling, filtering, pagination, uploads, metadata, versioning, and OpenAPI generation.

## Public Types

| Area | Types |
| --- | --- |
| Request/response | `Request`, `Response`, `APIError` |
| Serializers | `Serializer`, `SerializerField`, `FieldOptions`, `ModelSerializer`, `ModelSerializerConfig`, `OrderedFieldError` |
| Views | `APIView`, `View`, `RequestInitializer`, `RequestHook`, `ExceptionHandler`, `ResponseFinalizer` |
| ViewSets/routers | `ModelViewSet`, `ModelViewSetStore`, `ViewSetAction`, `Router`, `RouterOption`, `Route` |
| Authentication | `Token`, `AuthenticationResult`, `Authenticator`, `AuthenticatorFunc`, `TokenStore`, `MemoryTokenStore` |
| Permissions | `PermissionClass` |
| Throttling | `Rate`, `ThrottleError`, `ThrottleDecision`, `Throttle`, `ThrottleStore`, `MemoryThrottleStore`, `RateThrottle` |
| Filtering | `FilterBackend`, `FilterBackendFunc`, `FilterSet` |
| Pagination | `PaginatedResult`, `PageNumberPagination`, `LimitOffsetPagination`, `CursorPagination` |
| Parsers/uploads | `UploadedFile`, `MultipartBody`, `ParserRegistry`, `StoredUpload`, `UploadStorage`, `UploadConfig`, `UploadHandler`, `MemoryUploadStorage` |
| Renderers | `Renderer`, `JSONRenderer`, `BrowsableAPIRenderer`, `PlainTextRenderer` |
| Metadata | `MetadataOptions`, `APIMetadata`, `RouteMetadata`, `ActionMetadata`, `SerializerMetadata`, `SerializerFieldMetadata`, `FilterMetadata` |
| Versioning | `VersioningConfig`, `VersioningStrategy`, `URLPathVersioning`, `NamespaceVersioning`, `HostNameVersioning`, `QueryParameterVersioning`, `AcceptHeaderVersioning` |
| OpenAPI | `OpenAPIOptions`, `OpenAPISpec`, `OpenAPIInfo`, `OpenAPIExternalDocs`, `OpenAPIOperation`, `OperationMetadata`, `ResponseSchema`, `OpenAPIParameter`, `OpenAPIRequestBody`, `OpenAPIMediaType`, `OpenAPIComponents` |

## Serializer Fields

Serializer field constructors:

`BooleanField`, `IntegerField`, `FloatField`, `DecimalField`, `StringField`, `EmailField`, `URLField`, `SlugField`, `UUIDField`, `DateField`, `DateTimeField`, `TimeField`, `DurationField`, `JSONField`, `ChoiceField`, `ListField`, `DictField`, `NestedField`, and `MethodField`.

`FieldOptions` supports required, allow null, allow blank, default, source, label, help text, readonly, write-only, validators, and custom error messages.

## Authentication And Permissions

Authenticators:

- `SessionAuthentication`
- `TokenAuthentication`
- `AuthenticatorFunc`

Permission classes:

- `AllowAny`
- `IsAuthenticated`
- `IsAdminUser`
- `IsAuthenticatedOrReadOnly`
- `ModelPermissions`
- `CustomPermission`
- `CustomObjectPermission`

## Parsers And Renderers

Parsers support JSON, form, multipart, and raw request bodies through `ParserRegistry`.

Renderers include JSON, browsable API HTML, and plain text.

## Filtering, Pagination, Throttling

Filtering supports exact field lookups, search, ordering, custom filter backends, distinct behavior, and validation of exposed fields.

Pagination supports page-number, limit/offset, and cursor strategies.

Throttling supports request rate parsing, in-memory throttle storage, scoped keys, user/IP keys, retry-after decisions, and permission-style request hooks.

## OpenAPI

OpenAPI generation uses router, viewset, serializer, auth, permission, and metadata definitions to build paths, operations, request bodies, responses, security schemes, and components. `OpenAPIJSONView` serves a spec as an API view.

`Router.HandleHTTP` registers an existing `net/http.Handler` on the API router while keeping a named route and optional `OperationMetadata` for OpenAPI. API path parameters are available through `request.PathValue("name")` for raw handlers and through `Request.PathParam("name")` for Gogo-native API views.

`WithExceptionHandler` configures router-level handling for 404, 405, write failures, and uncaught API view panics. Use it when an existing public API must keep a legacy error body during migration. `WithTrailingSlash` keeps trailing-slash behavior configurable per API router, and nested routers preserve their registered route patterns.

## Errors

`ErrParse`, `ErrUnsupportedMediaType`, `ErrAuthenticationFailed`, `ErrNotAuthenticated`, `ErrPermissionDenied`, `ErrThrottled`, `ErrValidation`, and upload/storage errors where returned by upload handlers.

## Example

```go
serializer := api.NewSerializer(api.StringField("title", api.FieldOptions{Required: true}))
cleaned, errors, ok := serializer.Validate(map[string]any{"title": "Hello"})
_, _, _ = cleaned, errors, ok
```
