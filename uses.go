package rbac

import "webtyp.com/model"

// GrantsForUse translates a ModuleUse into the RBAC grants for every entity of def that allows operations.
func GrantsForUse(def model.ModuleDefinition, use model.ModuleUse) []model.Grant {
	var mask model.Action
	switch use {
	case model.ManagesEverything:
		mask = model.AllActions
	case model.WorksWithAllRecords, model.SeesOwnRecords:
		mask = model.Create | model.Read | model.Update
	case model.VisitsWithoutAccount:
		mask = model.Read
	default:
		return nil
	}
	var grants []model.Grant
	for i := range def.Entities {
		e := &def.Entities[i]
		actions := e.Allow & mask
		if actions == 0 {
			continue
		}
		res := model.ResourceOf(e)
		if res == "" {
			continue
		}
		grants = append(grants, model.Grant{Resource: res, Actions: actions})
	}
	return grants
}
