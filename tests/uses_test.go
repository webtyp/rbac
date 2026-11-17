package tests

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/rbac"
)

func TestUses_GrantsForUseCoversEachMode(t *testing.T) {
	def := model.ModuleDefinition{
		Name: "client_directory",
		Entities: []model.Definition{
			{
				Name:  "client",
				Allow: model.AllActions,
				Fields: model.Fields{
					{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
				},
			},
			{
				Name:  "internal_note",
				Allow: model.Read,
				Fields: model.Fields{
					{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
				},
			},
		},
	}

	admin := rbac.GrantsForUse(def, model.ManagesEverything)
	if len(admin) != 2 || admin[0].Actions != model.AllActions || admin[1].Actions != model.Read {
		t.Fatalf("unexpected admin grants: %+v", admin)
	}

	staff := rbac.GrantsForUse(def, model.WorksWithAllRecords)
	if len(staff) != 2 || staff[0].Actions != (model.Create|model.Read|model.Update) {
		t.Fatalf("unexpected staff grants: %+v", staff)
	}

	client := rbac.GrantsForUse(def, model.SeesOwnRecords)
	if len(client) != 2 || client[0].Actions != (model.Create|model.Read|model.Update) {
		t.Fatalf("unexpected client grants: %+v", client)
	}

	visitor := rbac.GrantsForUse(def, model.VisitsWithoutAccount)
	if len(visitor) != 2 || visitor[0].Actions != model.Read {
		t.Fatalf("unexpected visitor grants: %+v", visitor)
	}

	if none := rbac.GrantsForUse(def, 0); none != nil {
		t.Fatalf("expected nil for zero ModuleUse, got %+v", none)
	}
}
