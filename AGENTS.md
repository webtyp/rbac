# AGENTS.md — webtyp/rbac

Constraints for agents changing this library. Read before touching any file.

## What this repo is

Role-based authorization runtime (roles, grants, permission checks) for WebTyp applications. Authentication and sessions belong to `webtyp/auth`.

## It compiles to WASM

Its files carry no build tags: they compile for the server **and** for `GOOS=js GOARCH=wasm` (TinyGo). Code that reaches the browser binary follows TinyGo's constraints:

- **No `map`** in code compiled to wasm: use a slice with a linear scan, or `[]fmt.KeyValue`.
- **No reflection, ever**: no `reflect`, no `errors.Is` / `errors.As`, no `sort.Slice`, and no
  `==` / `!=` / `switch` between interface values with non-nil operands. Under TinyGo each of them
  pulls `internal/reflectlite` into the binary (`err == ErrX` included: `error` is an interface).
  Detect sentinels with their `IsX(err)` function (`orm.IsNotFound`, `storage.IsNoRows`).
- Strings and conversions through `webtyp.com/fmt`, not the standard library's `fmt`/`strconv`.

## Reuse before writing

Persistence goes through `webtyp.com/orm` / `webtyp.com/storage` (backend-agnostic `storage.Conn`);
never import a concrete backend (`sqlite`, `postgres`, `indexdb`) outside tests. A missing piece
in a base library is fixed there, never re-implemented here.

## The build that defines "done"

```bash
go install webtyp.com/devflow/cmd/gotest@latest   # once
gotest
```

## Rules

- Tests live in `tests/`. A root-level test is allowed only with a top-of-file comment justifying the unexported identifier it needs. Never export a symbol so a test can reach it.
- Every repeated string (error text, keys) is a named constant.
