# Auth Reference

The auth package provides Django-style users, groups, permissions, password hashing, authentication, sessions integration, forms, decorators, admin registration metadata, and password reset tokens.

## Public Types

| Area | Types |
| --- | --- |
| Models | `Permission`, `Group`, `AbstractBaseUser`, `AbstractUser`, `User` |
| Content types | `ContentType`, `ContentTypeRegistry` |
| Authentication | `Credentials`, `UserStore`, `MemoryUserStore`, `FileUserStore`, `UserIDLoader` |
| Passwords | `PasswordHasher`, `PBKDF2SHA256Hasher`, `Argon2IDHasher` |
| Forms | `LoginForm`, `PasswordChangeForm`, `PasswordResetRequestForm`, `PasswordResetConfirmForm`, `SetPasswordForm`, `UserCreationForm`, `UserChangeForm` |
| Tokens | `PasswordResetTokenSigner` |
| Admin | `AdminFieldset`, `AdminRegistration` |

## Models

`Permission`, `Group`, and `User` expose Django-compatible model metadata:

- `auth_permission`
- `auth_group`
- `auth_user`

`AbstractBaseUser` and `AbstractUser` provide embeddable metadata for custom inheritable user types.

## Permissions

Default model permissions are `add`, `change`, `delete`, and `view`.

Permission helpers:

- `GenerateModelPermissions`
- `HasPerm`
- `HasModulePerms`
- `GetUserPermissions`
- `GetGroupPermissions`
- `GetAllPermissions`
- `WithPerm`

Inactive users fail permission checks. Superusers pass all permission checks when active.

## Authentication

Use `Authenticate` with a `UserStore`. `MemoryUserStore` is useful for tests
and local examples. `FileUserStore` persists users as JSON and is compatible
with the generated project admin store at `.gogo/auth_users.json`.

`AuthenticationMiddleware` reads `user_id` from the request session and attaches either an authenticated user or `AnonymousUser` to the context.

Context helpers:

- `ContextWithUser`
- `UserFromContext`
- `AnonymousUser`

## Passwords

Password hashers:

- `PBKDF2SHA256Hasher`
- `Argon2IDHasher`

Password helpers cover encode, verify, harden runtime, must-update checks, and password validation through auth forms.

## Forms

Auth forms validate login, password change, password reset request, password reset confirmation, set password, user creation, and user change flows. `ErrFormValidation` reports validation failure.

## Decorators

HTTP decorators:

- `LoginRequired`
- `PermissionRequired`
- `UserPassesTest`
- `StaffMemberRequired`
- `SuperuserRequired`

## Tokens

`PasswordResetTokenSigner` signs and verifies password reset tokens. `ErrInvalidPasswordResetToken` identifies invalid or expired reset tokens.

## Error Types

`ErrInvalidPassword`, `ErrUnknownPasswordHasher`, `ErrInvalidPasswordResetToken`, and `ErrFormValidation`.

## Example

```go
store, err := auth.NewMemoryUserStore(auth.User{AbstractUser: auth.AbstractUser{Username: "admin"}})
user, ok, err := auth.Authenticate(context.Background(), store, auth.Credentials{Username: "admin"})
_, _, _ = user, ok, err
```
