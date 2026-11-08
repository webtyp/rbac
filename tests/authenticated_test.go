package tests

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/rbac"
)

func TestAuthenticated_AnySubjectHoldsIt(t *testing.T) {
	svc := newTestService(t)
	p := testProject

	grants := []model.Grant{
		{Resource: "chat_message", Actions: model.Create},
	}
	if err := svc.EnsureGrants(p, rbac.RoleAuthenticated, grants); err != nil {
		t.Fatalf("EnsureGrants: %v", err)
	}

	if !svc.Can(p, "u1", "chat_message", model.Create) {
		t.Errorf("expected subject 'u1' to hold RoleAuthenticated grant for Create")
	}
	if svc.Can(p, "u1", "chat_message", model.Delete) {
		t.Errorf("expected subject 'u1' to NOT hold ungranted Delete action")
	}
}

func TestAuthenticated_EmptySubjectDenied(t *testing.T) {
	svc := newTestService(t)
	p := testProject

	grants := []model.Grant{
		{Resource: "chat_message", Actions: model.Create},
	}
	if err := svc.EnsureGrants(p, rbac.RoleAuthenticated, grants); err != nil {
		t.Fatalf("EnsureGrants: %v", err)
	}

	if svc.Can(p, "", "chat_message", model.Create) {
		t.Errorf("expected empty subject to be denied")
	}
}

func TestAuthenticated_EnsureIdempotent(t *testing.T) {
	svc := newTestService(t)
	p := testProject

	grants := []model.Grant{
		{Resource: "chat_message", Actions: model.Create},
	}
	if err := svc.EnsureGrants(p, rbac.RoleAuthenticated, grants); err != nil {
		t.Fatalf("EnsureGrants 1: %v", err)
	}
	if err := svc.EnsureGrants(p, rbac.RoleAuthenticated, grants); err != nil {
		t.Fatalf("EnsureGrants 2: %v", err)
	}

	role, err := svc.GetRoleByCode(p, rbac.RoleAuthenticated)
	if err != nil {
		t.Fatalf("GetRoleByCode: %v", err)
	}
	if role == nil {
		t.Fatal("expected role to exist")
	}

	perm, err := svc.GetPermission(p, "chat_message:"+model.Create.String())
	if err != nil {
		t.Fatalf("GetPermission: %v", err)
	}
	if perm == nil {
		t.Fatal("expected perm to exist")
	}

	grantsRes, err := svc.SubjectGrants(p, "u1")
	if err != nil {
		t.Fatalf("SubjectGrants: %v", err)
	}
	if len(grantsRes) != 1 {
		t.Fatalf("expected exactly 1 grant, got %d", len(grantsRes))
	}
}

func TestAuthenticated_CacheInvalidated(t *testing.T) {
	svc := newTestService(t)
	p := testProject

	if svc.Can(p, "u1", "chat_message", model.Create) {
		t.Fatal("expected Can to be false before EnsureGrants")
	}

	grants := []model.Grant{
		{Resource: "chat_message", Actions: model.Create},
	}
	if err := svc.EnsureGrants(p, rbac.RoleAuthenticated, grants); err != nil {
		t.Fatalf("EnsureGrants: %v", err)
	}

	if !svc.Can(p, "u1", "chat_message", model.Create) {
		t.Fatal("expected Can to be true after EnsureGrants (cache invalidated)")
	}
}
