package conformance

import (
	"testing"

	"webtyp.com/ddl"
	"webtyp.com/orm"
	"webtyp.com/rbac"
	"webtyp.com/storage"
)

// Factory builds the database a clause runs on.
type Factory struct {
	Name string
	New  func(t *testing.T) (conn storage.Conn, compiler ddl.Compiler)
}

// Run runs the clauses every database must pass to serve rbac. A driver calls it from its own tests.
//
//	conformance.Run(t, conformance.Factory{Name: "sqlite", New: newSqliteDatabase})
func Run(t *testing.T, f Factory) {
	t.Run(f.Name, func(t *testing.T) {
		t.Run("migrate_detects_preexisting_duplicates", func(t *testing.T) {
			conn, ddlCompiler := f.New(t)
			db := orm.New(conn)
			if err := ddl.New(conn, ddlCompiler).Sync(&rbac.Role{}); err != nil {
				t.Fatal(err)
			}

			// Seed duplicate role rows directly (bypassing CreateRole, before the unique index exists)
			if err := db.Create(&rbac.Role{ProjectId: "proj-dup", Id: "r1", Code: "dupcode", Name: "N1"}); err != nil {
				t.Fatalf("db.Create r1: %v", err)
			}
			if err := db.Create(&rbac.Role{ProjectId: "proj-dup", Id: "r2", Code: "dupcode", Name: "N2"}); err != nil {
				t.Fatalf("db.Create r2: %v", err)
			}

			dups, err := rbac.FindDuplicateRoleCodes(db)
			if err != nil {
				t.Fatalf("FindDuplicateRoleCodes: %v", err)
			}
			if len(dups) != 1 || dups[0].ProjectID != "proj-dup" || dups[0].Code != "dupcode" {
				t.Fatalf("FindDuplicateRoleCodes expected [proj-dup, dupcode], got %+v", dups)
			}

			err = rbac.Migrate(db.RawConn(), ddlCompiler)
			if !rbac.IsDuplicateRoleCode(err) {
				t.Fatalf("expected Migrate to return ErrDuplicateRoleCode on preexisting duplicates, got %v", err)
			}
		})
	})
}
