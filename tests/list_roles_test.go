package tests

import (
	"testing"
)

func TestListRoles_OrderedByCode(t *testing.T) {
	svc := newTestService(t)
	if err := svc.CreateRole("p1", "r2", "zeta", "Z", ""); err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if err := svc.CreateRole("p1", "r1", "alpha", "A", ""); err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	roles, err := svc.ListRoles("p1")
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}

	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	if roles[0].Code != "alpha" || roles[1].Code != "zeta" {
		t.Errorf("expected codes [alpha, zeta] in that order, got [%s, %s]", roles[0].Code, roles[1].Code)
	}
}

func TestListRoles_ScopedByProject(t *testing.T) {
	svc := newTestService(t)
	if err := svc.CreateRole("p1", "r1", "p1_role", "P1", ""); err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if err := svc.CreateRole("p2", "r2", "p2_role", "P2", ""); err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	roles, err := svc.ListRoles("p2")
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}

	if len(roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(roles))
	}
	if roles[0].Code != "p2_role" {
		t.Errorf("expected p2_role, got %s", roles[0].Code)
	}
}

func TestListRoles_Empty(t *testing.T) {
	svc := newTestService(t)
	roles, err := svc.ListRoles("p1")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("expected len 0, got %d", len(roles))
	}
}
