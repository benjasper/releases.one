package migrate

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/schema"
)

//go:embed schema.hcl
var desiredSchemaHCL []byte

// Run brings the connected database up to date with schema.sql by applying only additive changes.
// Destructive changes (dropping tables, columns, indexes or foreign keys) are skipped and must be done by hand.
func Run(ctx context.Context, db *sql.DB) error {
	driver, err := mysql.Open(db)
	if err != nil {
		return err
	}

	locker, ok := driver.(schema.Locker)
	if !ok {
		return errors.New("migration driver does not support locking")
	}

	unlock, err := locker.Lock(ctx, "releases_one_migrate", time.Minute)
	if err != nil {
		return errors.Join(err, errors.New("failed to acquire migration lock"))
	}
	defer unlock()

	current, err := driver.InspectSchema(ctx, "", nil)
	if err != nil {
		return errors.Join(err, errors.New("failed to inspect database schema"))
	}

	var desired schema.Realm
	if err := mysql.EvalHCLBytes(desiredSchemaHCL, &desired, nil); err != nil {
		return errors.Join(err, errors.New("failed to parse desired schema"))
	}
	if len(desired.Schemas) != 1 {
		return fmt.Errorf("expected exactly one schema in schema.hcl, found %d", len(desired.Schemas))
	}

	// schema.hcl is generated against a dev database with a different name, so align it with the live database
	desiredSchema := desired.Schemas[0]
	desiredSchema.Name = current.Name

	changes, err := driver.SchemaDiff(current, desiredSchema, schema.DiffSkipChanges(
		&schema.ModifySchema{},
		&schema.DropSchema{},
		&schema.DropTable{},
		&schema.DropColumn{},
		&schema.DropIndex{},
		&schema.DropForeignKey{},
	))
	if err != nil {
		return errors.Join(err, errors.New("failed to diff database schema"))
	}

	if len(changes) == 0 {
		slog.Info("Database schema is up to date")
		return nil
	}

	plan, err := driver.PlanChanges(ctx, "auto_migrate", changes)
	if err != nil {
		return errors.Join(err, errors.New("failed to plan schema migration"))
	}

	for _, change := range plan.Changes {
		slog.Info(fmt.Sprintf("Applying schema migration: %s", change.Cmd))
	}

	if err := driver.ApplyChanges(ctx, changes); err != nil {
		return errors.Join(err, errors.New("failed to apply schema migration"))
	}

	slog.Info(fmt.Sprintf("Applied %d schema migration statements", len(plan.Changes)))

	return nil
}
