package rbac

import "webtyp.com/fmt"

// domainError is the concrete type of this package's sentinel errors. Code
// compares them by asserting this type and comparing the value: == between two
// error values compiles, under TinyGo, to runtime.interfaceEqual, which pulls
// internal/reflectlite into the wasm binary.
type domainError string

func (e domainError) Error() string { return string(e) }

// ErrNotFound reports that a role or permission id has no matching row.
const ErrNotFound domainError = "rbac not found"

// ErrDuplicateRoleCode reporta que la base tiene dos roles con el mismo code
// dentro de un proyecto — un estado que este paquete ya no permite crear pero
// que una base anterior a esta versión pudo haber acumulado. Se resuelve a
// mano: hay que decidir cuál de los dos roles sobrevive y reasignar sus
// usuarios. Migrate NO lo resuelve solo porque elegir cuál borrar es una
// decisión de política, no de esquema.
const ErrDuplicateRoleCode domainError = "rbac duplicate role code"

// ErrRoleNotFound reporta que el rol identificado por su code no existe en el
// proyecto especificado.
const ErrRoleNotFound domainError = "rbac role not found"

func IsNotFound(err error) bool {
	e, ok := err.(domainError)
	return ok && e == ErrNotFound
}

func IsDuplicateRoleCode(err error) bool {
	e, ok := err.(domainError)
	return ok && e == ErrDuplicateRoleCode
}

func IsRoleNotFound(err error) bool {
	e, ok := err.(domainError)
	return ok && e == ErrRoleNotFound
}

func isUniqueViolation(err error) bool {
	return fmt.Contains(err.Error(), "UNIQUE constraint failed") ||
		fmt.Contains(err.Error(), "constraint: unique") ||
		fmt.Contains(err.Error(), "duplicate key")
}
