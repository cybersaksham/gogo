# Auth Permissions And Sessions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the built-in, non-optional auth toolkit with inheritable users, groups, permissions, content types, sessions, password management, middleware, decorators, and admin/API integration.

**Architecture:** Public package `auth` owns identity, password, permission, and authentication APIs. Public package `sessions` owns session storage and middleware. Auth models are framework-provided and mandatory; applications extend user data through embedding and profile-style extension models without replacing the framework-owned auth user table in this product version.

**Tech Stack:** Models, ORM, migrations, secure password hashing, cookies, CSRF integration hooks, context-aware middleware.

---

## Files

- Create: `auth/models.go`
- Create: `auth/contenttypes.go`
- Create: `auth/permissions.go`
- Create: `auth/password.go`
- Create: `auth/tokens.go`
- Create: `auth/authenticate.go`
- Create: `auth/middleware.go`
- Create: `auth/decorators.go`
- Create: `auth/forms.go`
- Create: `auth/admin.go`
- Create: `auth/migrations/0001_initial.go`
- Create: `sessions/session.go`
- Create: `sessions/store.go`
- Create: `sessions/middleware.go`
- Create: `sessions/cookie.go`
- Modify: `internal/cli/auth.go`
- Create: `auth/models_test.go`
- Create: `auth/password_test.go`
- Create: `auth/permissions_test.go`
- Create: `sessions/session_test.go`

## Task 1: Add Auth Models

- [ ] Create `auth/models.go`.
- [ ] Define `Permission` with fields:
  - ID
  - Name
  - ContentTypeID
  - Codename
- [ ] Define `Group` with fields:
  - ID
  - Name
  - Permissions many-to-many
- [ ] Define `User` with fields:
  - ID
  - Password
  - LastLogin
  - IsSuperuser
  - Username
  - FirstName
  - LastName
  - Email
  - IsStaff
  - IsActive
  - DateJoined
  - Groups many-to-many
  - UserPermissions many-to-many
- [ ] Define `AbstractUser` and `AbstractBaseUser` embeddable structs.
- [ ] Add metadata matching default permissions and admin names.
- [ ] Add tests for metadata, fields, relationships, and inheritance.
- [ ] Run `go test ./auth`.
- [ ] Commit with message `Add Built In Auth Models`.

## Task 2: Add Content Types

- [ ] Create `auth/contenttypes.go`.
- [ ] Define `ContentType` with app label, model name, and natural key.
- [ ] Generate content types from model registry during migrations.
- [ ] Support lookup by model, natural key, and ID.
- [ ] Add tests for creation, lookup, duplicate prevention, and stale content type detection.
- [ ] Run `go test ./auth`.
- [ ] Commit with message `Add Content Type Registry`.

## Task 3: Add Permission Generation And Checks

- [ ] Create `auth/permissions.go`.
- [ ] Generate default permissions for every model:
  - `add`
  - `change`
  - `delete`
  - `view`
- [ ] Support custom model permissions from metadata.
- [ ] Implement:
  - `HasPerm(user, "app.codename")`
  - `HasModulePerms(user, "app")`
  - `GetUserPermissions`
  - `GetGroupPermissions`
  - `GetAllPermissions`
  - `WithPerm`
- [ ] Superusers must pass all permission checks when active.
- [ ] Inactive users must fail non-anonymous permission checks.
- [ ] Add tests for direct permissions, group permissions, superuser, inactive user, module perms, and custom permissions.
- [ ] Run `go test ./auth`.
- [ ] Commit with message `Add Auth Permission Checks`.

## Task 4: Add Password Hashing

- [ ] Create `auth/password.go`.
- [ ] Support password hashers:
  - PBKDF2-SHA256
  - Argon2id
  - BCrypt if dependency is approved during implementation
- [ ] Store algorithm, iterations or parameters, salt, and hash.
- [ ] Implement:
  - `MakePassword`
  - `CheckPassword`
  - `IsPasswordUsable`
  - `SetUnusablePassword`
  - `MustUpdatePasswordHash`
- [ ] Add password validators:
  - Minimum length
  - Common password rejection
  - Numeric password rejection
  - User attribute similarity
- [ ] Add tests with known vectors, invalid hashes, unusable passwords, parameter upgrades, and validators.
- [ ] Run `go test ./auth`.
- [ ] Commit with message `Add Password Hashing`.

## Task 5: Add Authentication Backend

- [ ] Create `auth/authenticate.go`.
- [ ] Implement built-in username/email authentication.
- [ ] Do not make auth backend configurable in this product version.
- [ ] Normalize username and email.
- [ ] Reject inactive users.
- [ ] Update last login through a controlled service.
- [ ] Add tests for username login, email login, invalid password, inactive user, and last login update.
- [ ] Run `go test ./auth`.
- [ ] Commit with message `Add Built In Authentication`.

## Task 6: Add Session Model And Store

- [ ] Create `sessions/session.go`.
- [ ] Create `sessions/store.go`.
- [ ] Define session table fields:
  - Session key
  - Session data
  - Expire date
- [ ] Implement signed session keys and storage backends:
  - Database-backed sessions
  - Cached database sessions
  - Cache-only sessions
  - File-backed sessions
  - Signed cookie sessions
- [ ] Support create, load, save, delete, cycle key, flush, expiry age, expiry date, and modified/accessed tracking.
- [ ] Add tests for session creation, expiry, tamper detection, cycle key, flush, save behavior, and every session backend.
- [ ] Run `go test ./sessions`.
- [ ] Commit with message `Add Server Side Sessions`.

## Task 7: Add Session And Auth Middleware

- [ ] Create `sessions/middleware.go`.
- [ ] Create `auth/middleware.go`.
- [ ] Session middleware must attach session to request context.
- [ ] Auth middleware must attach user and anonymous user to request context.
- [ ] Support secure cookie attributes:
  - HttpOnly
  - Secure
  - SameSite
  - Path
  - Domain
  - MaxAge
- [ ] Add tests for middleware order, anonymous request, authenticated request, expired session, and secure cookie settings.
- [ ] Run `go test ./sessions ./auth`.
- [ ] Commit with message `Add Session And Auth Middleware`.

## Task 8: Add Login Logout Password Flows

- [ ] Create `auth/forms.go`.
- [ ] Implement:
  - Login form
  - Password change form
  - Password reset request form
  - Password reset confirm form
  - Set password form
  - User creation form
  - User change form
- [ ] Create `auth/tokens.go`.
- [ ] Implement signed password reset tokens with expiry.
- [ ] Add tests for form validation, password validators, reset token validity, expired token, and changed-password invalidation.
- [ ] Run `go test ./auth`.
- [ ] Commit with message `Add Auth Forms And Tokens`.

## Task 9: Add Access Helpers

- [ ] Create `auth/decorators.go`.
- [ ] Implement:
  - `LoginRequired`
  - `PermissionRequired`
  - `UserPassesTest`
  - `StaffRequired`
  - `SuperuserRequired`
- [ ] Support redirect to login for HTML requests and 401/403 for API requests.
- [ ] Add tests for each helper.
- [ ] Run `go test ./auth`.
- [ ] Commit with message `Add Auth Access Helpers`.

## Task 10: Add Auth Migrations

- [ ] Create `auth/migrations/0001_initial.go`.
- [ ] Create tables:
  - `gogo_content_type`
  - `auth_permission`
  - `auth_group`
  - `auth_group_permissions`
  - `auth_user`
  - `auth_user_groups`
  - `auth_user_user_permissions`
  - `gogo_session`
- [ ] Add indexes and unique constraints matching lookup behavior.
- [ ] Add migration tests using migration executor.
- [ ] Run `go test ./auth ./sessions ./migrations`.
- [ ] Commit with message `Add Auth And Session Migrations`.

## Task 11: Add Auth CLI Commands

- [ ] Modify `internal/cli/auth.go`.
- [ ] Implement `createsuperuser` with non-interactive flags:
  - `--username`
  - `--email`
  - `--password`
  - `--database`
- [ ] Implement `changepassword`.
- [ ] Refuse password through command history in interactive mode unless explicitly provided for automation.
- [ ] Add tests for user creation, duplicate user, invalid password, and password change.
- [ ] Run `go test ./internal/cli ./auth`.
- [ ] Commit with message `Add Auth CLI Commands`.

## Task 12: Add Auth Admin Registration

- [ ] Create `auth/admin.go`.
- [ ] Register users, groups, permissions, content types, and sessions with admin.
- [ ] Define list displays, filters, search fields, fieldsets, readonly fields, and actions.
- [ ] Add tests after admin phase exists; before admin phase, compile-test metadata adapters.
- [ ] Run `go test ./auth`.
- [ ] Commit with message `Add Auth Admin Registration`.

## Acceptance Checklist

- [ ] User, group, permission, content type, and session tables exist.
- [ ] Auth models are extendable through embedding and extension models without replacing the framework-owned auth user table.
- [ ] Default and custom permissions are generated.
- [ ] Password hashing supports secure upgrades.
- [ ] Sessions are signed and support database, cached database, cache, file, and signed-cookie storage backends.
- [ ] Auth middleware attaches anonymous or authenticated users.
- [ ] CLI can create superusers and change passwords.
- [ ] Admin and API phases can consume auth permissions.
