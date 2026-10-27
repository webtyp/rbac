---
PLAN: "feat: SubjectGrants — a subject's grants as stored, one wildcard rule for server and browser"
TAG: v0.0.22
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

# Plan — `webtyp.com/rbac`: `SubjectGrants`

## Why

The browser decides which screens to draw from the signed-in user's permissions (the `me` response of
`webtyp/auth`). Today the server translates the admin's wildcard grant (`*`) into one row per resource of a list
every app writes by hand; a screen missing from that list is hidden even from the admin. The fix: send the grants
**as stored** (`*` included) and let the browser apply the same rule the server applies, `model.Grant.Matches` /
`model.AnyGrant`. This repo hands out a subject's grants in that typed form.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
```

## Ecosystem search

- `model.Grant{Resource, Actions}`, `Grant.Matches`, `model.AnyGrant` (`webtyp.com/model`, `rbac.go:198-225`):
  the typed grant and THE wildcard rule → **reuse**.
- `rbac` already resolves a subject's permissions (`subjectGrants`, `subject_grants.go`, cached in `ucache`) and
  converts each row in `HasPermission` (`rbac.go:372-397`) → **extend**: expose the converted list.
- Name: `Grants` is taken by `model.PolicyDescriber.Grants() []RoleGrant` (policy-wide, per role) → this method is
  `SubjectGrants` (per subject), matching the existing `CanSubject`/`subjectGrants` vocabulary.

## Design gate

1. **Prior art.** Casbin `GetImplicitPermissionsForUser` returns the stored rows, not an expansion; CASL sends
   `{action: 'manage', subject: 'all'}` to the client and its client-side `can()` understands it; Spring Security
   exposes a collection of `GrantedAuthority`. All hand out the grants as stored and evaluate them with one rule.
2. **Novice-name test.** `svc.SubjectGrants(projectID, subjectID)` — «the grants of this subject». Same first two
   parameters as `Can`.
3. **Complexity ledger.** Concepts +1 / −0; ways to evaluate a wildcard: −1 (`HasPermission` stops hand-converting
   rows and calls `model.AnyGrant`, the rule `model` declares as the only one).
4. **Where it belongs.** `rbac` owns the rows; `model` owns the rule; `auth` (another plan) consumes the list
   through its own port without importing `rbac`.
5. **What it deletes.** The per-row conversion loop inside `HasPermission` (it now reuses `SubjectGrants`). The old
   unexecuted `docs/PLAN_ALLOWED_ACTIONS.md` was already deleted.

## APIs used

| Symbol | Exact signature | Tag | File |
|---|---|---|---|
| `model.Grant` | `struct { Resource Resource; Actions Action }` | model v0.2.6 | `rbac.go` |
| `model.AnyGrant` | `func AnyGrant(grants []Grant, r Resource, a Action) bool` — empty slice denies | model v0.2.6 | `rbac.go` |
| `model.ParseAction` | `func ParseAction(s string) (Action, error)` | model v0.2.6 | `rbac.go` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — `SubjectGrants`, and `HasPermission` on top of it

**File:** `rbac.go`. Add, right before `HasPermission`:

```go
// SubjectGrants returns every grant subjectID holds within projectID, as stored:
// a wildcard grant stays {Resource: model.Wildcard}, never expanded. Evaluate the
// result with model.AnyGrant — the same rule HasPermission applies — so a server
// and a browser agree on what "*" means. An empty or unknown subject has no
// grants (nil, nil). A stored action that is not a CRUD string is an error, never
// a silently dropped grant.
func (m *Service) SubjectGrants(projectID, subjectID string) ([]model.Grant, error) {
	if subjectID == "" {
		return nil, nil
	}
	g, err := m.subjectGrants(projectID, subjectID)
	if err != nil {
		return nil, err
	}
	grants := make([]model.Grant, 0, len(g.Permissions))
	for _, p := range g.Permissions {
		actions, err := model.ParseAction(p.Action)
		if err != nil {
			return nil, err
		}
		grants = append(grants, model.Grant{Resource: model.Resource(p.Resource), Actions: actions})
	}
	return grants, nil
}
```

Replace the WHOLE body of `HasPermission` with:

```go
	grants, err := m.SubjectGrants(projectID, subjectID)
	if err != nil {
		return false, err
	}
	return model.AnyGrant(grants, resource, action), nil
```

Keep its doc comment. Check: `grep -n "grant.Matches\|model.Grant{" rbac.go` → only inside `SubjectGrants`.

## Stage 2 — tests

**File:** new `tests/subject_grants_test.go`, using `newTestService`/`newTestServiceWithDB` and `testProject` from
`tests/rbac_test.go` (same seeding calls as `TestFullFlow`: `CreateRole`, `CreatePermission`, `AssignPermission`,
`AssignRole`).

| Test | Setup | Call | Expected |
|---|---|---|---|
| `TestSubjectGrants_Empty` | nothing | `SubjectGrants(testProject, "nobody")` | `len == 0`, `err == nil` |
| `TestSubjectGrants_EmptySubject` | nothing | `SubjectGrants(testProject, "")` | `nil`, `nil` |
| `TestSubjectGrants_AsStored` | role with `invoice` Read and `report` Read+Update, assigned to `user-1` | `SubjectGrants` | contains `{invoice, model.Read}` and `{report, model.Read\|model.Update}` |
| `TestSubjectGrants_WildcardNotExpanded` | role with permission resource `model.Wildcard`, `model.AllActions` | `SubjectGrants` | exactly one grant, `Resource == model.Wildcard`, `Actions == model.AllActions`; and `model.AnyGrant(grants, "anything", model.Delete) == true` |
| `TestSubjectGrants_ProjectIsolated` | same role in project A only | `SubjectGrants("B", "user-1")` | empty |
| `TestSubjectGrants_CorruptActionFails` | seed like `TestHasPermission_CorruptActionFailsLoudly` (`Action: "raed"`) | `SubjectGrants` | `err != nil` |

Existing tests (`TestFullFlow`, `TestClosedByDefault`, `TestHasPermission_CorruptActionFailsLoudly`,
`TestAdminRevokesAccess_TakesEffectImmediately`) must stay green unchanged: they prove `HasPermission` kept its behaviour.

## Stage 3 — docs

- `README.md`, table «API — Quiero X → Uso Y»: add the row
  `| Permisos de un sujeto tal como están guardados (para el navegador) | \`svc.SubjectGrants(projectID, subjectID) ([]model.Grant, error)\` — evaluar con \`model.AnyGrant\` |`.
- `docs/ARCHITECTURE.md`: one paragraph beside `Can`/`HasPermission`: `SubjectGrants` returns the rows unexpanded;
  `HasPermission` is `model.AnyGrant` over them, so there is one wildcard rule.

## Delivery

- `gotest` green. A PR that does not compile, or whose tests could not run, is a failed delivery.
  Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`.
- `grep -rn "webtyp.com/auth" go.mod` → empty (no new dependency).
- Never pause to ask permission or report progress: when the check passes, commit, push and open the PR. Stuck 30
  minutes → open the PR listing what fails. Commit after every stage; never `git reset`, `git restore`,
  `git checkout -- <file>` or `git stash` on your own work.

| Stage | File | Action |
|---|---|---|
| 1 | `rbac.go` | add `SubjectGrants`; `HasPermission` = `model.AnyGrant` over it |
| 2 | `tests/subject_grants_test.go` | grant matrix |
| 3 | `README.md`, `docs/ARCHITECTURE.md` | document |
