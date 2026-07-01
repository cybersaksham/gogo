# Admin Reference

The admin package provides model registration, sites, staff access policies, auth views, index pages, change lists, change forms, delete confirmation, history, filters, search, autocomplete, widgets, inlines, actions, static assets, and queue/admin docs integrations.

## Public Types

| Area | Types |
| --- | --- |
| Site | `Site`, `SiteOptions`, `SiteCollection`, `PermissionPolicy`, `StaffPermissionPolicy`, `SessionPermissionPolicy` |
| Registry | `Registry`, `ModelAdmin`, `ModelAdminHooks` |
| Model options | `Fieldset`, `Inline`, `InlineKind`, `URLPattern`, `ComputedColumn` |
| Auth views | `AuthViewConfig` |
| Change list | `ChangeList`, `ChangeListColumn`, `ChangeListRow`, `DateBucket` |
| Change form | `ChangeFormInput`, `ChangeFormContext`, `ChangeFormField`, `RelatedPopup`, `JavaScriptCatalogResponse` |
| Delete | `DeletionObject`, `DeletionSummary` |
| History | `HistoryPage`, `AdminLogEntry`, `MemoryLogStore` |
| Filters/search | `FilterChoice`, `FilterSpec`, `FilterState`, `FilterResult`, `SearchOptions`, `SearchQuery`, `AutocompleteConfig`, `AutocompleteResult`, `AutocompleteResponse` |
| Actions | `Action`, `ActionContext`, `ActionResult`, `ActionStore`, `ActionRegistry`, `MemoryActionStore` |
| Inlines | `InlineInput`, `InlineFormset`, `InlineForm`, `InlineStore`, `MemoryInlineStore` |
| Widgets | `WidgetChoice`, `WidgetConfig` |

## ModelAdmin Options

`ModelAdmin` supports:

`Actions`, `ActionsOnTop`, `ActionsOnBottom`, `ActionsSelectionCounter`, `AutocompleteFields`, `DateHierarchy`, `EmptyValueDisplay`, `Exclude`, `Fields`, `Fieldsets`, `FilterHorizontal`, `FilterVertical`, `Form`, `FormfieldOverrides`, `Inlines`, `ListDisplay`, `ListDisplayLinks`, `ListEditable`, `ListFilter`, `ListMaxShowAll`, `ListPerPage`, `ListSelectRelated`, `Ordering`, `Paginator`, `PrepopulatedFields`, `PreserveFilters`, `RadioFields`, `RawIDFields`, `ReadonlyFields`, `ReadOnly`, `SaveAs`, `SaveAsContinue`, `SaveOnTop`, `SearchFields`, `SearchHelpText`, `ShowFacets`, `SortableBy`, `ViewOnSite`, `CustomURLs`, `ComputedColumns`, `ActionDefinitions`, and `Hooks`.

Set `ReadOnly` to true for admin registrations that should allow staff users to
view modules and objects while blocking add, change, and delete by default.

## Hooks

`ModelAdminHooks` covers queryset, ordering, search, display, filters, readonly fields, fields, fieldsets, forms, save/delete lifecycle, related saves, response hooks, messaging, lookup checks, deleted object summaries, paginator, autocomplete, prepopulation, select-related, sortable fields, inlines, custom URLs, and add/change/delete/view/module permissions.

## Views

Admin auth views:

- `LoginView`
- `LogoutView`
- `PasswordChangeView`
- `PasswordChangeDoneView`

Admin site access defaults to active authenticated staff users. Generated
projects configure `SessionPermissionPolicy` with the built-in file user store
and file session store so `/admin/` redirects anonymous users to login and
allows staff users created with `go run manage.go createsuperuser`.

## Errors

`ErrInvalidURLPrefix`, `ErrDuplicateSite`, `ErrAlreadyRegistered`, `ErrNotRegistered`, `ErrUnmanagedModel`, `ErrInvalidModelAdminOption`, `ErrAdminPermissionDenied`, `ErrInvalidChangeListQuery`, `ErrProtectedRelation`, and `ErrInvalidInlineFormset`.

## Example

```go
meta := models.Metadata{AppLabel: "blog", ModelName: "Post", TableName: "blog_post"}
registry := admin.NewRegistry()
err := registry.RegisterMetadata(meta, admin.ModelAdmin{ListDisplay: []string{"title"}})
_ = err
```
