package operations

import (
	"testing"

	"github.com/cybersaksham/gogo/migrations"
	"github.com/cybersaksham/gogo/models"
)

func TestOperationSpecRoundTripsDatabaseDefault(t *testing.T) {
	defaultValue := models.DefaultSQL("now()")
	spec := migrations.OperationSpecFor(AddField{
		AppLabel:  "blog",
		ModelName: "Post",
		TableName: "blog_post",
		Field: migrations.FieldState{
			Name:      "created_at",
			Column:    "created_at",
			Kind:      "timestamp",
			DBDefault: &defaultValue,
		},
		HasDefault: true,
	})

	data, err := spec.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON() error = %v", err)
	}
	decoded, err := migrations.OperationSpecFromJSON(data)
	if err != nil {
		t.Fatalf("OperationSpecFromJSON() error = %v", err)
	}
	if decoded.Field == nil || decoded.Field.DBDefault == nil || decoded.Field.DBDefault.Kind != models.DefaultExpression || decoded.Field.DBDefault.SQL != "now()" {
		t.Fatalf("decoded default = %#v", decoded.Field)
	}
}
