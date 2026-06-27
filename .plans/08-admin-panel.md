# Admin Panel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Django-style built-in admin panel with model registration, CRUD screens, permissions, filters, search, actions, inlines, widgets, templates, static assets, history, and customization hooks.

**Architecture:** Public package `admin` exposes site and model-admin configuration. The admin uses framework HTTP, auth, forms, ORM, templates, static files, and model metadata while keeping generated URLs and templates overridable.

**Tech Stack:** Gogo HTTP router, auth permissions, forms, templates, ORM querysets, embedded static assets, server-side HTML.

---

## Files

- Create: `admin/site.go`
- Create: `admin/registry.go`
- Create: `admin/model_admin.go`
- Create: `admin/options.go`
- Create: `admin/urls.go`
- Create: `admin/views.go`
- Create: `admin/change_list.go`
- Create: `admin/change_form.go`
- Create: `admin/delete.go`
- Create: `admin/history.go`
- Create: `admin/actions.go`
- Create: `admin/filters.go`
- Create: `admin/search.go`
- Create: `admin/inlines.go`
- Create: `admin/widgets.go`
- Create: `admin/forms.go`
- Create: `admin/permissions.go`
- Create: `admin/log.go`
- Create: `admin/templates/base.html`
- Create: `admin/templates/index.html`
- Create: `admin/templates/login.html`
- Create: `admin/templates/change_list.html`
- Create: `admin/templates/change_form.html`
- Create: `admin/templates/delete_confirmation.html`
- Create: `admin/templates/history.html`
- Create: `admin/static/admin.css`
- Create: `admin/static/admin.js`
- Create: `admin/site_test.go`
- Create: `admin/model_admin_test.go`
- Create: `admin/views_test.go`

## Task 1: Implement Admin Site

- [ ] Create `admin/site.go`.
- [ ] Define `Site` with:
  - Name
  - Header
  - Title
  - Index title
  - URL prefix
  - Login view
  - Logout view
  - Password change view
  - Permission policy
  - Model registry
- [ ] Support default site and multiple named sites.
- [ ] Add tests for default site, custom site, multiple sites, and URL prefix validation.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Site`.

## Task 2: Implement Model Registration

- [ ] Create `admin/registry.go`.
- [ ] Implement `Register(model, adminConfig)`, `Unregister(model)`, `IsRegistered(model)`, and `GetAdmin(model)`.
- [ ] Reject duplicate registrations and unmanaged models unless explicitly allowed.
- [ ] Add autodiscovery hook through app registry.
- [ ] Add tests for registration, duplicate registration, unregister, and autodiscovery order.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Model Registry`.

## Task 3: Implement ModelAdmin Options

- [ ] Create `admin/model_admin.go`.
- [ ] Create `admin/options.go`.
- [ ] Support options:
  - `Actions`
  - `ActionsOnTop`
  - `ActionsOnBottom`
  - `ActionsSelectionCounter`
  - `AutocompleteFields`
  - `DateHierarchy`
  - `EmptyValueDisplay`
  - `Exclude`
  - `Fields`
  - `Fieldsets`
  - `FilterHorizontal`
  - `FilterVertical`
  - `Form`
  - `FormfieldOverrides`
  - `Inlines`
  - `ListDisplay`
  - `ListDisplayLinks`
  - `ListEditable`
  - `ListFilter`
  - `ListMaxShowAll`
  - `ListPerPage`
  - `ListSelectRelated`
  - `Ordering`
  - `Paginator`
  - `PrepopulatedFields`
  - `PreserveFilters`
  - `RadioFields`
  - `RawIDFields`
  - `ReadonlyFields`
  - `SaveAs`
  - `SaveAsContinue`
  - `SaveOnTop`
  - `SearchFields`
  - `SearchHelpText`
  - `ShowFacets`
  - `SortableBy`
  - `ViewOnSite`
- [ ] Support per-request admin URL extension through `GetURLs`.
- [ ] Support hooks:
  - `GetQuerySet`
  - `GetOrdering`
  - `GetSearchResults`
  - `GetListDisplay`
  - `GetListFilter`
  - `GetReadonlyFields`
  - `GetFields`
  - `GetFieldsets`
  - `GetForm`
  - `SaveModel`
  - `SaveForm`
  - `SaveFormset`
  - `DeleteModel`
  - `DeleteQueryset`
  - `SaveRelated`
  - `ResponseAdd`
  - `ResponseChange`
  - `ResponseDelete`
  - `MessageUser`
  - `LookupAllowed`
  - `GetDeletedObjects`
  - `GetChangeList`
  - `GetPaginator`
  - `GetAutocompleteFields`
  - `GetPrepopulatedFields`
  - `GetListSelectRelated`
  - `GetSortableBy`
  - `GetInlineInstances`
  - `GetInlines`
  - `HasAddPermission`
  - `HasChangePermission`
  - `HasDeletePermission`
  - `HasViewPermission`
  - `HasModulePermission`
- [ ] Add validation tests for incompatible options such as editable fields missing from list display.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add ModelAdmin Options`.

## Task 4: Implement Admin URLs

- [ ] Create `admin/urls.go`.
- [ ] Generate namespaced routes:
  - `admin:index`
  - `admin:login`
  - `admin:logout`
  - `admin:password_change`
  - `admin:app_list`
  - `admin:<app>_<model>_changelist`
  - `admin:<app>_<model>_add`
  - `admin:<app>_<model>_change`
  - `admin:<app>_<model>_delete`
  - `admin:<app>_<model>_history`
  - `admin:<app>_<model>_autocomplete`
  - `admin:<app>_<model>_jsi18n`
  - Custom routes returned by `GetURLs`
- [ ] Add route reversing tests.
- [ ] Run `go test ./admin ./http`.
- [ ] Commit with message `Add Admin URLs`.

## Task 5: Implement Admin Authentication Views

- [ ] Create `admin/views.go`.
- [ ] Implement login, logout, password change, and password change done views.
- [ ] Require staff status for admin access.
- [ ] Preserve safe next URLs.
- [ ] Block open redirects.
- [ ] Add tests for staff login, non-staff denial, inactive denial, next URL safety, and logout.
- [ ] Run `go test ./admin ./auth`.
- [ ] Commit with message `Add Admin Auth Views`.

## Task 6: Implement Admin Index And App List

- [ ] Render admin index grouped by app label.
- [ ] Show only models the user can view or change.
- [ ] Include add/change URLs based on permissions.
- [ ] Add tests for permission-filtered app list.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Index`.

## Task 7: Implement Change List

- [ ] Create `admin/change_list.go`.
- [ ] Support:
  - List display columns
  - Computed columns
  - Boolean icons
  - Empty values
  - Sorting
  - Pagination
  - Show all limit
  - List editable
  - Bulk selection
  - Preserve filters
  - Date hierarchy
  - Query parameter validation
  - Popup selection responses
  - Preserved filters after add, change, and delete
- [ ] Add tests for every option and invalid query parameter handling.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Change List`.

## Task 8: Implement Search And Filters

- [ ] Create `admin/search.go`.
- [ ] Create `admin/filters.go`.
- [ ] Support search prefixes:
  - Exact
  - Case-insensitive contains
  - Full text hook for supported dialects
  - Related field traversal
- [ ] Implement autocomplete JSON endpoint with permission checks, search fields, pagination, and forwarded field constraints.
- [ ] Support filters:
  - Boolean filter
  - Choices filter
  - Date filter
  - Related object filter
  - Empty field filter
  - Custom simple list filter
  - Facet counts when enabled
- [ ] Add tests for search SQL, filter choices, facets, and related filters.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Search And Filters`.

## Task 9: Implement Change Form

- [ ] Create `admin/change_form.go`.
- [ ] Support add and edit forms.
- [ ] Support fieldsets, readonly fields, prepopulated fields, raw ID fields, autocomplete fields, radio fields, horizontal and vertical many-to-many filters.
- [ ] Support save, save and continue, save and add another, save as new, and delete.
- [ ] Support admin popups for adding and selecting related objects.
- [ ] Support JavaScript translation catalog route for admin widgets.
- [ ] Add tests for form rendering metadata, validation, permission checks, and save buttons.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Change Form`.

## Task 10: Implement Inlines

- [ ] Create `admin/inlines.go`.
- [ ] Support:
  - Stacked inline
  - Tabular inline
  - Extra forms
  - Min forms
  - Max forms
  - Can delete
  - Show change link
  - FK name selection
  - Inline permissions
- [ ] Add tests for inline formsets, validation, saving, deletion, and permission filtering.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Inlines`.

## Task 11: Implement Actions

- [ ] Create `admin/actions.go`.
- [ ] Add built-in delete selected action.
- [ ] Support custom actions with confirmation pages.
- [ ] Support global actions and per-model actions.
- [ ] Support action permissions.
- [ ] Add tests for action execution, confirmation, permission checks, and action errors.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Actions`.

## Task 12: Implement Widgets

- [ ] Create `admin/widgets.go`.
- [ ] Implement:
  - Text input
  - Textarea
  - Number input
  - Checkbox
  - Select
  - Select multiple
  - Date input
  - Time input
  - DateTime input
  - File input
  - Clearable file input
  - Raw ID relation widget
  - Autocomplete widget
  - Filtered select multiple
  - Readonly display widget
- [ ] Add tests for rendered attributes, escaping, selected values, and relation URLs.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Widgets`.

## Task 13: Implement Delete And History

- [ ] Create `admin/delete.go`.
- [ ] Create `admin/history.go`.
- [ ] Create `admin/log.go`.
- [ ] Add `AdminLogEntry` model with action time, user, content type, object ID, object repr, action flag, and change message.
- [ ] Show deletion collector summary before deletion.
- [ ] Log additions, changes, deletions, and actions.
- [ ] Add tests for delete confirmation, protected relation handling, log entries, and history page rendering.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Delete And History`.

## Task 14: Implement Templates And Static Assets

- [ ] Create all admin templates and static files listed in this plan.
- [ ] Provide semantic HTML, accessible form labels, keyboard-friendly tables, and responsive layout.
- [ ] Do not require client projects to copy assets.
- [ ] Allow template override by project template directories.
- [ ] Add tests that embedded assets exist and templates render required blocks.
- [ ] Run `go test ./admin`.
- [ ] Commit with message `Add Admin Templates And Assets`.

## Acceptance Checklist

- [ ] Admin supports model CRUD with permission checks.
- [ ] Admin supports list display, search, filters, pagination, sorting, fieldsets, readonly fields, widgets, actions, inlines, history, and delete confirmation.
- [ ] Auth models are registered in admin.
- [ ] Admin URLs are namespaced and reversible.
- [ ] Admin templates are overridable.
- [ ] Staff-only access is enforced.
