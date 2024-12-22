package dataaccess

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/jackc/tern/v2/migrate"
)

const versionTable = "db_version"

type Migrator struct {
	migrator *migrate.Migrator
}

//go:embed migrations/*.sql
var migrationFiles embed.FS

func NewMigrator() (Migrator, error) {
	conn, err := getConnection()

	if err != nil {
		return Migrator{}, err
	}

	migrator, err := migrate.NewMigrator(context.Background(), conn, versionTable)
	if err != nil {
		return Migrator{}, err
	}

	migrationRoot, _ := fs.Sub(migrationFiles, "migrations")

	err = migrator.LoadMigrations(migrationRoot)
	if err != nil {
		return Migrator{}, err
	}

	return Migrator{
		migrator: migrator,
	}, nil
}

// Info the current migration version and the embedded maximum migration, and a textual
// representation of the migration state for informational purposes.
func (m Migrator) Info() (int32, int32, string, error) {

	version, err := m.migrator.GetCurrentVersion(context.Background())
	if err != nil {
		return 0, 0, "", err
	}
	info := ""

	var last int32
	for _, thisMigration := range m.migrator.Migrations {
		last = thisMigration.Sequence

		cur := version == thisMigration.Sequence
		indicator := "  "
		if cur {
			indicator = "->"
		}
		info = info + fmt.Sprintf(
			"%2s %3d %s\n",
			indicator,
			thisMigration.Sequence, thisMigration.Name)
	}

	return version, last, info, nil
}

// Migrate migrates the DB to the most recent version of the schema.
func (m Migrator) Migrate() error {
	err := m.migrator.Migrate(context.Background())
	return err
}

// MigrateTo migrates to a specific version of the schema. Use '0' to undo all migrations.
func (m Migrator) MigrateTo(ver int32) error {
	err := m.migrator.MigrateTo(context.Background(), ver)
	return err
}
