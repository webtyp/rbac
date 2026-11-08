package rbac

import (
	"webtyp.com/model"
	"webtyp.com/orm"
)

// RoleAuthenticated is held by every subject with a non-empty id, without a UserRole row. Grant it what any signed-in
// person may do; a module declares those grants (DefaultGrants) and the app applies them with EnsureGrants.
const RoleAuthenticated model.RoleCode = "authenticated"

// EnsureGrants makes role exist in projectID (id = its code, name = its code) and gives it every grant, creating the
// permissions it needs. Idempotent: running it again changes nothing.
//
//	err := svc.EnsureGrants(app.TenantID, rbac.RoleAuthenticated, chatroom.DefaultGrants)
func (m *Service) EnsureGrants(projectID string, role model.RoleCode, grants []model.Grant) error {
	r, err := m.GetRoleByCode(projectID, role)
	if err != nil {
		if IsRoleNotFound(err) || orm.IsNotFound(err) {
			if err := m.CreateRole(projectID, string(role), role, string(role), ""); err != nil {
				return err
			}
			r, err = m.GetRoleByCode(projectID, role)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	roleID := r.Id
	actions := []model.Action{model.Create, model.Read, model.Update, model.Delete}

	for _, grant := range grants {
		for _, action := range actions {
			if !grant.Actions.Has(action) {
				continue
			}
			permID := string(grant.Resource) + ":" + action.String()
			_, err := m.GetPermission(projectID, permID)
			if err != nil {
				if IsRoleNotFound(err) || orm.IsNotFound(err) {
					if err := m.CreatePermission(projectID, permID, permID, grant.Resource, action); err != nil {
						return err
					}
				} else {
					return err
				}
			}
			if err := m.AssignPermission(projectID, roleID, permID); err != nil {
				return err
			}
		}
	}

	m.ucache.InvalidateByRole(roleID)
	if role == RoleAuthenticated {
		m.ucache.Clear()
	}
	return nil
}
