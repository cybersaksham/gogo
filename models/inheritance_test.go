package models

import (
	"errors"
	"reflect"
	"testing"
)

func TestAbstractBaseFieldsAreInheritedAndOverridden(t *testing.T) {
	base := Metadata{
		AppLabel:  "core",
		ModelName: "Timestamped",
		Abstract:  true,
		Fields: []FieldMeta{
			{Name: "created_at", Column: "created_at"},
			{Name: "slug", Column: "base_slug"},
		},
	}
	child := Metadata{
		AppLabel:  "blog",
		ModelName: "Post",
		Fields: []FieldMeta{
			{Name: "slug", Column: "post_slug"},
			{Name: "title", Column: "title"},
		},
	}

	resolved, err := ResolveInheritance(child, WithAbstractBase(base))
	if err != nil {
		t.Fatalf("ResolveInheritance() error = %v", err)
	}
	if names := fieldNames(resolved.Fields); !reflect.DeepEqual(names, []string{"created_at", "slug", "title"}) {
		t.Fatalf("fields = %#v", names)
	}
	if resolved.Fields[0].SourceModel != "core.Timestamped" {
		t.Fatalf("created_at source = %q", resolved.Fields[0].SourceModel)
	}
	if resolved.Fields[1].Column != "post_slug" || resolved.Fields[1].SourceModel != "blog.Post" {
		t.Fatalf("overridden slug field = %#v", resolved.Fields[1])
	}
	if len(resolved.Inheritance.AbstractBases) != 1 || resolved.Inheritance.AbstractBases[0].Label() != "core.Timestamped" {
		t.Fatalf("abstract base metadata = %#v", resolved.Inheritance.AbstractBases)
	}
}

func TestMultiTableInheritanceAddsParentLinksAndOrdering(t *testing.T) {
	parent := Metadata{AppLabel: "people", ModelName: "Person", TableName: "people_person"}
	child := Metadata{AppLabel: "people", ModelName: "Employee", TableName: "people_employee"}

	resolved, err := ResolveInheritance(child, WithMultiTableParent(parent))
	if err != nil {
		t.Fatalf("ResolveInheritance() error = %v", err)
	}
	if len(resolved.Inheritance.MultiTableParents) != 1 {
		t.Fatalf("parents = %#v", resolved.Inheritance.MultiTableParents)
	}
	link := resolved.Inheritance.MultiTableParents[0]
	if link.Parent.Label() != "people.Person" || link.Field.Name != "person_ptr" || !link.Field.ParentLink {
		t.Fatalf("parent link = %#v", link)
	}
	if got := ParentSaveOrder(resolved); !reflect.DeepEqual(got, []string{"people.Person", "people.Employee"}) {
		t.Fatalf("ParentSaveOrder() = %#v", got)
	}
	if got := ParentDeleteOrder(resolved, false); !reflect.DeepEqual(got, []string{"people.Employee", "people.Person"}) {
		t.Fatalf("ParentDeleteOrder(false) = %#v", got)
	}
	if got := ParentDeleteOrder(resolved, true); !reflect.DeepEqual(got, []string{"people.Employee"}) {
		t.Fatalf("ParentDeleteOrder(true) = %#v", got)
	}
}

func TestProxyInheritanceUsesParentTableAndMetadata(t *testing.T) {
	parent := Metadata{
		AppLabel:           "orders",
		ModelName:          "Order",
		TableName:          "orders_order",
		DefaultManagerName: "objects",
		Fields:             []FieldMeta{{Name: "status"}},
	}
	proxy := Metadata{
		AppLabel:           "orders",
		ModelName:          "OpenOrder",
		Proxy:              true,
		DefaultManagerName: "open_objects",
	}

	resolved, err := ResolveInheritance(proxy, WithProxyBase(parent))
	if err != nil {
		t.Fatalf("ResolveInheritance() error = %v", err)
	}
	if resolved.TableName != "orders_order" || resolved.DBTable != "orders_order" {
		t.Fatalf("proxy table = (%q, %q)", resolved.TableName, resolved.DBTable)
	}
	if resolved.Inheritance.ProxyFor == nil || resolved.Inheritance.ProxyFor.Label() != "orders.Order" {
		t.Fatalf("proxy metadata = %#v", resolved.Inheritance.ProxyFor)
	}
	if resolved.DefaultManagerName != "open_objects" {
		t.Fatalf("DefaultManagerName = %q", resolved.DefaultManagerName)
	}
	if names := fieldNames(resolved.Fields); !reflect.DeepEqual(names, []string{"status"}) {
		t.Fatalf("proxy fields = %#v", names)
	}
}

func TestAuthUserExtensionMetadataPreservesFrameworkUserTable(t *testing.T) {
	extension := Metadata{AppLabel: "accounts", ModelName: "CustomerUserExtension"}
	resolved, err := ResolveInheritance(extension, WithAuthUserExtension(AuthUserExtension{
		UserModel:       ModelRef{AppLabel: "auth", ModelName: "User", TableName: "auth_user"},
		ProfileRelation: "profile",
		ExtensionFields: []FieldMeta{{Name: "timezone"}},
	}))
	if err != nil {
		t.Fatalf("ResolveInheritance() error = %v", err)
	}
	auth := resolved.Inheritance.AuthUserExtension
	if auth == nil || auth.UserModel.TableName != "auth_user" || !auth.PreservesFrameworkUserTable {
		t.Fatalf("auth extension metadata = %#v", auth)
	}
	if auth.ProfileRelation != "profile" || resolved.Fields[0].Name != "timezone" {
		t.Fatalf("auth extension fields/relation = %#v / %#v", auth, resolved.Fields)
	}
}

func TestInheritanceValidationFailures(t *testing.T) {
	parent := Metadata{AppLabel: "bad", ModelName: "Parent"}
	_, err := ResolveInheritance(Metadata{AppLabel: "bad", ModelName: "Child"}, WithAbstractBase(parent))
	if !errors.Is(err, ErrInvalidMetadata) {
		t.Fatalf("abstract parent error = %v, want ErrInvalidMetadata", err)
	}

	_, err = ResolveInheritance(Metadata{AppLabel: "bad", ModelName: "Proxy"}, WithProxyBase(Metadata{AppLabel: "bad", ModelName: "Base"}))
	if !errors.Is(err, ErrInvalidMetadata) {
		t.Fatalf("proxy error = %v, want ErrInvalidMetadata", err)
	}

	_, err = ResolveInheritance(Metadata{AppLabel: "bad", ModelName: "Ext"}, WithAuthUserExtension(AuthUserExtension{}))
	if !errors.Is(err, ErrInvalidMetadata) {
		t.Fatalf("auth extension error = %v, want ErrInvalidMetadata", err)
	}
}

func fieldNames(fields []FieldMeta) []string {
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = field.Name
	}
	return names
}
