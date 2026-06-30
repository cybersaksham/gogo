package auth

import (
	"time"

	"github.com/cybersaksham/gogo/models"
)

const (
	appLabel              = "auth"
	permissionModel       = "Permission"
	groupModel            = "Group"
	userModel             = "User"
	permissionTable       = "auth_permission"
	groupTable            = "auth_group"
	userTable             = "auth_user"
	permissionTarget      = "auth.Permission"
	groupTarget           = "auth.Group"
	contentTypeTarget     = "auth.ContentType"
	defaultManagerName    = "objects"
	defaultBaseManager    = "objects"
	defaultPermissionName = "permission"
	defaultGroupName      = "group"
	defaultUserName       = "user"
)

// Permission grants one model-level capability for a content type.
type Permission struct {
	ID            int64
	Name          string
	ContentTypeID int64
	Codename      string
	ContentType   ContentType
	AppLabel      string
}

// ModelMeta returns Django-compatible metadata for auth permissions.
func (Permission) ModelMeta() models.Metadata {
	return authMetadata(models.Metadata{
		AppLabel:          appLabel,
		ModelName:         permissionModel,
		TableName:         permissionTable,
		DBTable:           permissionTable,
		VerboseName:       defaultPermissionName,
		VerboseNamePlural: "permissions",
		Fields: []models.FieldMeta{
			{Name: "id", Column: "id", PrimaryKey: true},
			{Name: "name", Column: "name"},
			{Name: "content_type", Column: "content_type_id", RelationTarget: contentTypeTarget, DeleteBehavior: "cascade"},
			{Name: "codename", Column: "codename"},
		},
		Constraints: []models.Constraint{
			{Name: "auth_permission_content_type_codename_uniq", Type: models.ConstraintUnique, Fields: []models.IndexField{models.Asc("content_type_id"), models.Asc("codename")}},
		},
	})
}

// Group stores a named collection of permissions.
type Group struct {
	ID          int64
	Name        string
	Permissions []Permission
}

// ModelMeta returns Django-compatible metadata for auth groups.
func (Group) ModelMeta() models.Metadata {
	return authMetadata(models.Metadata{
		AppLabel:          appLabel,
		ModelName:         groupModel,
		TableName:         groupTable,
		DBTable:           groupTable,
		VerboseName:       defaultGroupName,
		VerboseNamePlural: "groups",
		Fields: []models.FieldMeta{
			{Name: "id", Column: "id", PrimaryKey: true},
			{Name: "name", Column: "name"},
			{Name: "permissions", RelationTarget: permissionTarget},
		},
		Constraints: []models.Constraint{
			{Name: "auth_group_name_uniq", Type: models.ConstraintUnique, Fields: []models.IndexField{models.Asc("name")}},
		},
	})
}

// AbstractBaseUser contains the core identity fields shared by user types.
type AbstractBaseUser struct {
	ID              int64
	Password        string
	LastLogin       time.Time
	IsSuperuser     bool
	IsActive        bool
	DateJoined      time.Time
	Groups          []Group
	Permissions     []Permission
	UserPermissions []Permission
	Anonymous       bool
	Authenticated   bool
}

// IsAuthenticated reports whether this principal represents a logged-in user.
func (u AbstractBaseUser) IsAuthenticated() bool {
	return u.Authenticated || (!u.Anonymous && u.ID != 0)
}

// IsAnonymous reports whether this principal represents an anonymous request.
func (u AbstractBaseUser) IsAnonymous() bool {
	return u.Anonymous
}

// AbstractUser adds the default Django username, contact, and staff fields.
type AbstractUser struct {
	AbstractBaseUser
	Username  string
	FirstName string
	LastName  string
	Email     string
	IsStaff   bool
}

// User is the framework-owned built-in user model.
type User struct {
	AbstractUser
}

// ModelMeta returns Django-compatible metadata for auth users.
func (User) ModelMeta() models.Metadata {
	return userMetadata(false)
}

// AbstractBaseUserMetadata exposes the embeddable base-user metadata.
func AbstractBaseUserMetadata() models.Metadata {
	meta := userMetadata(true)
	meta.ModelName = "AbstractBaseUser"
	meta.TableName = ""
	meta.DBTable = ""
	meta.VerboseName = "abstract base user"
	meta.VerboseNamePlural = "abstract base users"
	meta.Fields = meta.Fields[:4]
	return meta
}

// AbstractUserMetadata exposes the embeddable default-user metadata.
func AbstractUserMetadata() models.Metadata {
	meta := userMetadata(true)
	meta.ModelName = "AbstractUser"
	meta.TableName = ""
	meta.DBTable = ""
	meta.VerboseName = "abstract user"
	meta.VerboseNamePlural = "abstract users"
	return meta
}

// ModelMetadata returns the framework-owned auth models that client projects
// should include by default.
func ModelMetadata() []models.Metadata {
	return []models.Metadata{
		ContentType{}.ModelMeta(),
		Permission{}.ModelMeta(),
		Group{}.ModelMeta(),
		User{}.ModelMeta(),
	}
}

func userMetadata(abstract bool) models.Metadata {
	return authMetadata(models.Metadata{
		AppLabel:          appLabel,
		ModelName:         userModel,
		TableName:         userTable,
		DBTable:           userTable,
		VerboseName:       defaultUserName,
		VerboseNamePlural: "users",
		Abstract:          abstract,
		Fields: []models.FieldMeta{
			{Name: "id", Column: "id", PrimaryKey: true},
			{Name: "password", Column: "password"},
			{Name: "last_login", Column: "last_login"},
			{Name: "is_superuser", Column: "is_superuser"},
			{Name: "username", Column: "username"},
			{Name: "first_name", Column: "first_name"},
			{Name: "last_name", Column: "last_name"},
			{Name: "email", Column: "email"},
			{Name: "is_staff", Column: "is_staff"},
			{Name: "is_active", Column: "is_active"},
			{Name: "date_joined", Column: "date_joined"},
			{Name: "groups", RelationTarget: groupTarget},
			{Name: "user_permissions", RelationTarget: permissionTarget},
		},
		Constraints: []models.Constraint{
			{Name: "auth_user_username_uniq", Type: models.ConstraintUnique, Fields: []models.IndexField{models.Asc("username")}},
		},
	})
}

func authMetadata(meta models.Metadata) models.Metadata {
	meta.DefaultManagerName = defaultManagerName
	meta.BaseManagerName = defaultBaseManager
	meta.DefaultPermissions = []string{"add", "change", "delete", "view"}
	return meta
}
