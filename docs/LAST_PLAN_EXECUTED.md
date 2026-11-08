---
PLAN: "feat: implicit role authenticated — every signed-in subject holds it; EnsureGrants applies a module's default grants"
EXECUTOR: antigravity
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit after every stage and stop. Never ask whether to
> proceed. This plan runs **locally** and is **published as soon as the delivery check passes** (no `replace` anywhere).
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.

# Plan: the `authenticated` role

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
```

## Why

Decision D24 of `webtyp/docs/ARCHITECTURE_MASTER_PLAN.md`. Some actions belong to every signed-in person (write in the
chat, open a direct conversation, mark read). With crud every action requires a resource, and «anyone signed in» must stay
a grant in RBAC (data, adjustable per installation), never a flag in code. Prior art: Django REST `IsAuthenticated`
(https://www.django-rest-framework.org/api-guide/permissions/#isauthenticated), Keycloak default roles
(https://www.keycloak.org/docs/latest/server_admin/#_default_roles).

## Stage 1 — `authenticated.go` (new)

```go
// RoleAuthenticated is held by every subject with a non-empty id, without a UserRole row. Grant it what any signed-in
// person may do; a module declares those grants (DefaultGrants) and the app applies them with EnsureGrants.
const RoleAuthenticated model.RoleCode = "authenticated"

// EnsureGrants makes role exist in projectID (id = its code, name = its code) and gives it every grant, creating the
// permissions it needs. Idempotent: running it again changes nothing.
//
//	err := svc.EnsureGrants(app.TenantID, rbac.RoleAuthenticated, chatroom.DefaultGrants)
func (m *Service) EnsureGrants(projectID string, role model.RoleCode, grants []model.Grant) error
```
Body: `GetRoleByCode`; not found (`IsRoleNotFound`) → `CreateRole(projectID, string(role), role, string(role), "")`. For each grant and
each action bit in `grant.Actions` (`model.Create`, `Read`, `Update`, `Delete`): `permID := string(grant.Resource) + ":" + action.String()`
(the same id `registerRBAC` uses, `rbac.go:357`); create the permission if missing (`GetPermission` not found → `CreatePermission`) and
`AssignPermission(projectID, roleID, permID)` if not assigned yet. Then `m.ucache.InvalidateByRole(roleID)`.

## Stage 2 — `subject_grants.go`

In `resolveSubjectGrants`, for `subjectID != ""`: also load the role with code `RoleAuthenticated` of `projectID` (if it exists) and
add its id to `roleIDs` before reading roles and role permissions — even when the subject has no `UserRole` rows (today the
function returns early at `len(userRoles) == 0`; that early return now happens only when there are no user roles **and** no
authenticated role). An empty `subjectID` still gets nothing (`SubjectGrants` returns `nil, nil` first — unchanged).

## Stage 3 — tests (`tests/authenticated_test.go`, rows in `docs/TESTS.md`), `orm.New(mem.New())`

| Test | Expected |
|---|---|
| `TestAuthenticated_AnySubjectHoldsIt` | `EnsureGrants(p, RoleAuthenticated, [{chat_message, Create}])`; subject "u1" with no roles → `Can(p,"u1","chat_message",Create)` true |
| `TestAuthenticated_EmptySubjectDenied` | same setup → `Can(p,"","chat_message",Create)` false |
| `TestAuthenticated_EnsureIdempotent` | `EnsureGrants` twice → one role, one permission per action, no error |
| `TestAuthenticated_CacheInvalidated` | `Can` false, then `EnsureGrants` adds the grant → `Can` true (cache refreshed) |

## Stage 4 — documentation

`README.md`: «Anyone signed in» section with `RoleAuthenticated` and `EnsureGrants`. `docs/ARCHITECTURE.md`: how the implicit role
joins `resolveSubjectGrants`. `langc sync .` if any text changed.

## Delivery check

`gotest`. A tree that does not compile or whose tests did not run is a failed delivery.

| Stage | Files |
|---|---|
| 1 | `authenticated.go` |
| 2 | `subject_grants.go` |
| 3 | `tests/authenticated_test.go`, `docs/TESTS.md` |
| 4 | `README.md`, `docs/ARCHITECTURE.md` |
