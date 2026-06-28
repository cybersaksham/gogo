# Blog Tutorial

This tutorial expands the quickstart into a blog with relationships, forms, admin filters, paginated APIs, and a background email task.

## Domain Models

Create these models:

- `Author`: name, email, bio.
- `Post`: author foreign key, title, slug, body, status, published timestamp.
- `Tag`: name and slug.
- `Comment`: post foreign key, author name, email, body, approved flag.
- `AuditEvent`: actor, action, object label, timestamp.

Relationships:

- `Post` belongs to `Author`.
- `Post` has many `Tag` records through a many-to-many relation.
- `Comment` belongs to `Post`.
- `AuditEvent` records admin and API changes.

## Forms

Use `forms.NewForm` for public comment submission:

```go
commentForm := forms.NewForm(forms.FormOptions{
	Fields: map[string]*forms.Field{
		"name":  forms.NewCharField(forms.FieldOptions{Required: true}),
		"email": forms.NewEmailField(forms.FieldOptions{Required: true}),
		"body":  forms.NewCharField(forms.FieldOptions{Required: true}),
	},
})
```

Use model forms for admin-like create/update flows when the store can validate uniqueness.

## Admin

Register `Author`, `Post`, `Tag`, `Comment`, and `AuditEvent`.

For `Post`, configure:

```go
admin.ModelAdmin{
	ListDisplay:       []string{"title", "author", "status", "published_at"},
	ListFilter:        []string{"status", "author", "tags"},
	SearchFields:      []string{"title", "body", "author__name"},
	PrepopulatedFields: map[string][]string{"slug": {"title"}},
	ReadonlyFields:    []string{"created_at", "updated_at"},
}
```

`ListFilter` lets staff filter by status, author, and tags.

## API Pagination

Expose a post list with `api.PageNumberPagination`:

```go
paginator := api.PageNumberPagination{PageSize: 20}
page := paginator.Paginate(request, rows)
```

Use serializers for `Author`, `Post`, `Tag`, and `Comment`. Add filtering for status and tag slug, search on title/body, and ordering by published date.

## Background Email Task

Send an email after a comment is approved:

```go
_, _ = queueApp.RegisterTask("blog.email_comment_approved", func(ctx context.Context, args ...any) (any, error) {
	message := email.Message{Subject: "Comment approved", To: []string{args[0].(string)}}
	_, err := mailer.SendMessages(ctx, []email.Message{message})
	return nil, err
}, queue.TaskOptions{Queue: "email"})

signature := queue.NewSignature("blog.email_comment_approved", "reader@example.com").WithQueue("email")
```

The `email` package supplies messages and backends. `queue.NewSignature` creates the task call.

## Audit Trail

Create `AuditEvent` rows from admin save hooks, API create/update hooks, and background task completion hooks. Keep audit writes in the same transaction as the domain change when possible.

## Test The Flow

Use `testing.NewClient` for API tests, `testing.NewMailOutbox` for email assertions, `testing.NewQueueHarness` for eager queue tests, and fixture helpers for authors, posts, tags, and comments.
