---
PLAN: "feat: ListRoles — list a project's roles through the Service"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 763602360687102089
PR: https://github.com/webtyp/rbac/pull/4
REVIEWED_COMMIT: 6eb24217a7cf25e3be5c59427fd489ec2ad2c1aa
---

> Read AGENTS.md at the repo root before any change.

# PLAN — `Service.ListRoles`

## Why

`Service` can create, get, delete and assign roles, but cannot **list** the roles of a project.
`veltylabs/iam` worked around it by querying this package's `role` table from the app
(`db.Query(&rbac.Role{}).Where(rbac.Role_.ProjectId)…`), and the new `webtyp/access` needs the same
list. A missing method in the base is fixed here, once, not re-written by each consumer.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
```

## Design gate

1. **Prior art.** Casbin `GetAllRoles()`; Keycloak admin API `GET /roles`; Django `Group.objects.all()`
   — every RBAC store lists its roles; here scoped by project, like every other `Service` method.
2. **Novice-name test.** `svc.ListRoles(projectID)` — "list the roles of this project".
3. **Complexity ledger.** Concepts +0 · call-site lines 1 · ways to list roles: +1 / −1 (consumers stop
   querying the table directly).
4. **Where it belongs.** Here: the `role` table and its cache belong to this package.
5. **What it deletes.** The direct table query in consumers (`veltylabs/iam/config/admin_backend.go`,
   removed by that repo's own plan).

## Stage 1 — `rbac.go`

Add, next to `GetRoleByCode`:

```go
// ListRoles returns every role of projectID, ordered by code. A project with
// no roles returns an empty slice and a nil error.
//
// Example: roles, err := svc.ListRoles("mjosefa-cms")
func (m *Service) ListRoles(projectID string) ([]Role, error) {
	qb := m.db.Query(&Role{}).Where(Role_.ProjectId).Eq(projectID).OrderBy(Role_.Code).Asc()
	ptrs, err := ReadAllRole(qb)
	if err != nil {
		return nil, err
	}
	roles := make([]Role, len(ptrs))
	for i, r := range ptrs {
		roles[i] = *r
	}
	return roles, nil
}
```

## Stage 2 — `tests/list_roles_test.go` (new, `package tests`)

Use `newTestService(t)` from `tests/rbac_test.go`.

| Test | Setup → call | Expected |
|---|---|---|
| `TestListRoles_OrderedByCode` | `CreateRole("p1","r2","zeta","Z","")`, `CreateRole("p1","r1","alpha","A","")` → `ListRoles("p1")` | 2 roles, codes `alpha`, `zeta` in that order |
| `TestListRoles_ScopedByProject` | roles in `p1` and `p2` → `ListRoles("p2")` | only `p2`'s roles |
| `TestListRoles_Empty` | no roles → `ListRoles("p1")` | `len == 0`, `err == nil` |

## Stage 3 — docs

`README.md`: in the API list/table of `Service` methods, add `ListRoles(projectID) ([]Role, error)`
with one line: "every role of a project, ordered by code".

## Delivery check

```
GOPRIVATE=webtyp.com/*,github.com/veltylabs/* GOFLAGS=-mod=mod gotest
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never pause to ask permission or report progress: when the check passes, commit, push and open
  the PR. Stuck for 30 minutes: open the PR anyway listing what fails.
- Commit after every stage. Never run `git reset`, `git restore`, `git checkout -- <file>` or
  `git stash` on your own work.
- `git diff --stat main` lists: `rbac.go`, `tests/list_roles_test.go`, `README.md`, `AGENTS.md`, this plan.

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `rbac.go` | compiles |
| 2 | `tests/list_roles_test.go` | 3 tests pass |
| 3 | `README.md` | method listed |
