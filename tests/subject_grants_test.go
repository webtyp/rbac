package tests

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/rbac"
)

func TestSubjectGrants_Empty(t *testing.T) {
	svc := newTestService(t)
	grants, err := svc.SubjectGrants(testProject, "nobody")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(grants) != 0 {
		t.Fatalf("expected empty grants, got %v", grants)
	}
}

func TestSubjectGrants_EmptySubject(t *testing.T) {
	svc := newTestService(t)
	grants, err := svc.SubjectGrants(testProject, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if grants != nil {
		t.Fatalf("expected nil grants, got %v", grants)
	}
}

func TestSubjectGrants_AsStored(t *testing.T) {
	svc := newTestService(t)

	_ = svc.CreateRole(testProject, "role-1", "role-1", "Role 1", "")
	_ = svc.CreatePermission(testProject, "perm-1", "perm-1", "invoice", model.Read)
	_ = svc.CreatePermission(testProject, "perm-2", "perm-2", "report", model.Read|model.Update)
	_ = svc.AssignPermission(testProject, "role-1", "perm-1")
	_ = svc.AssignPermission(testProject, "role-1", "perm-2")
	_ = svc.AssignRole(testProject, "user-1", "role-1")

	grants, err := svc.SubjectGrants(testProject, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(grants) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(grants))
	}

	var hasInvoice, hasReport bool
	for _, g := range grants {
		if g.Resource == "invoice" && g.Actions == model.Read {
			hasInvoice = true
		}
		if g.Resource == "report" && g.Actions == (model.Read|model.Update) {
			hasReport = true
		}
	}

	if !hasInvoice || !hasReport {
		t.Fatalf("grants didn't match expected values: %v", grants)
	}
}

func TestSubjectGrants_WildcardNotExpanded(t *testing.T) {
	svc := newTestService(t)

	_ = svc.CreateRole(testProject, "role-1", "role-1", "Role 1", "")
	_ = svc.CreatePermission(testProject, "perm-1", "perm-1", model.Wildcard, model.AllActions)
	_ = svc.AssignPermission(testProject, "role-1", "perm-1")
	_ = svc.AssignRole(testProject, "user-1", "role-1")

	grants, err := svc.SubjectGrants(testProject, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(grants) != 1 {
		t.Fatalf("expected 1 grant, got %d", len(grants))
	}

	if grants[0].Resource != model.Wildcard || grants[0].Actions != model.AllActions {
		t.Fatalf("expected wildcard grant, got %v", grants[0])
	}

	if !model.AnyGrant(grants, "anything", model.Delete) {
		t.Fatalf("expected AnyGrant to allow 'anything' Delete")
	}
}

func TestSubjectGrants_ProjectIsolated(t *testing.T) {
	svc := newTestService(t)

	_ = svc.CreateRole(testProject, "role-1", "role-1", "Role 1", "")
	_ = svc.CreatePermission(testProject, "perm-1", "perm-1", "invoice", model.Read)
	_ = svc.AssignPermission(testProject, "role-1", "perm-1")
	_ = svc.AssignRole(testProject, "user-1", "role-1")

	grants, err := svc.SubjectGrants("B", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(grants) != 0 {
		t.Fatalf("expected empty grants for isolated project, got %v", grants)
	}
}

func TestSubjectGrants_CorruptActionFails(t *testing.T) {
	db, svc := newTestServiceWithDB(t)

	if err := svc.CreateRole(testProject, "r_corrupt", "editor", "Editor", ""); err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if err := db.Create(&rbac.Permission{
		ProjectId: testProject, Id: "p_corrupt", Name: "corrupt",
		Resource: "docs", Action: "raed", // not a CRUD string: ParseAction must fail
	}); err != nil {
		t.Fatalf("seed corrupt permission: %v", err)
	}
	if err := svc.AssignPermission(testProject, "r_corrupt", "p_corrupt"); err != nil {
		t.Fatalf("AssignPermission: %v", err)
	}
	if err := svc.AssignRole(testProject, "user-corrupt", "r_corrupt"); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}

	_, err := svc.SubjectGrants(testProject, "user-corrupt")
	if err == nil {
		t.Fatalf("expected error from corrupt action string, got nil")
	}
}
