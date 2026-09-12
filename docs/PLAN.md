---
PLAN: "feat: AllowedActions — the action set a subject holds on one resource"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Phase B** of
> [`LAN_RUT_AUTH_MASTER_PLAN.md`](https://github.com/tinywasm/app/blob/main/docs/LAN_RUT_AUTH_MASTER_PLAN.md).
> Parallel with phase A (`webtyp/auth`); the leaf app (phase D) waits for both
> tags. Doctrine: `CONSTRUCTION_HARNESS.md` in `tinywasm/app` docs.

# Plan — `webtyp.com/rbac`: one call per resource, not four

## 0. Context

`webtyp.com/auth` (phase A) adds the port:

```go
type ActionResolver interface {
    AllowedActions(projectID, subjectID string, resource model.Resource) model.Action
}
```

`rbac.Service` must satisfy it **structurally** — rbac never imports `auth`
(no dependency in either direction; the composition root injects
`rbac.Service` where `auth.ActionResolver` is asked for). Today a consumer
that needs the full action set of a subject on a resource must call
`Can(projectID, subjectID, resource, action)` once per action and OR the
results — a loop every consumer would write identically, i.e. glue that
belongs here.

## Design gate (api-design — five answers)

1. **Prior art.** **Casbin**: `get_implicit_permissions_for_user` returns the
   whole permission set per subject (one call, not one query per verb).
   **CanCanCan** (Rails): `accessible_by` / `can?` collapse ability checks
   into one object per resource. **Spring Security**: authorities are read as
   a collection per authentication, not probed one by one. All three answer
   "what may this subject do?" in one call; we differ only in the return
   type — the ecosystem's typed `model.Action` bitmask instead of strings —
   because `model` already closed the verb set (no invented verbs compile).
2. **Novice-name test.** `AllowedActions(projectID, subjectID, resource)`
   reads as "the actions that are allowed for this subject on this resource".
   It reuses the exact vocabulary of the existing `Can` (same first three
   parameters, minus the action) — nothing new to learn.
3. **Complexity ledger.** Concepts +1 / −0 (one method, same vocabulary).
   Call-site lines +1 / −5 (the four-`Can` loop disappears at every
   consumer). Ways to do the same thing +0 / −0 (`Can` stays: it answers the
   single-action question guards need; `AllowedActions` answers the
   set question profiles need — different intents, not two paths to one).
4. **Where it belongs.** `Service` owns grant resolution; the set form is the
   same concern aggregated, not a second concern. The port lives in `auth`
   (the consumer of the contract); the implementation lives here (the owner
   of the data). Neither imports the other.
5. **What it deletes.** The per-consumer 4×`Can` loop (in
   `mjosefa-cms/config/profile.go` today — deleted by that repo's phase D
   plan, not by this one).

## Stage 1 — the method

**File:** `module.go` (next to `Can`/`CanSubject`).

```go
// AllowedActions returns the set of CRUD actions subjectID holds on resource
// within projectID — the aggregate form of Can. The zero value means "no
// action granted" (closed by default). It satisfies auth.ActionResolver
// structurally; rbac does not import auth.
func (s *Service) AllowedActions(projectID, subjectID string, resource model.Resource) model.Action {
	var actions model.Action
	for _, a := range []model.Action{model.Create, model.Read, model.Update, model.Delete} {
		if s.Can(projectID, subjectID, resource, a) {
			actions |= a
		}
	}
	return actions
}
```

Implementation note: reuses `Can` (and therefore `HasPermission`'s existing
wildcard/role resolution and caches) — correctness over a hand-written query.
Do **not** reimplement grant resolution here.

## Stage 2 — tests

**File:** `tests/rbac_test.go` (extend) or a new `tests/allowed_actions_test.go`,
following the existing in-memory harness in `tests/`.

Cases: no grants → `0`; single grant → that action; wildcard role
(`model.Wildcard`, `model.AllActions` — the seeding pattern apps use) →
`model.AllActions`; mixed grants across two resources → each resource
resolves independently; `AllowedActions(...).String()` renders CRUD letters
(e.g. `model.Read|model.Update` → `"ru"`).

## Stage 3 — docs

`README.md`: one row in the method table + a 5-line example showing
`AllowedActions` feeding a profile-style response. `docs/ARCHITECTURE.md`:
mention the aggregate form next to `Can`. VERIFY against the implementation.

## Acceptance criteria

1. `go build ./...`, `go vet ./...`, `gotest ./...` green.
2. `grep -rn "webtyp.com/auth" go.mod` → empty (no dependency added).
3. The new test proves the wildcard path returns `model.AllActions`.

| Stage | File | Action |
|---|---|---|
| 1 | `module.go` | add `AllowedActions` |
| 2 | `tests/` | grant-matrix cases |
| 3 | `README.md`, `docs/ARCHITECTURE.md` | verify docs |
