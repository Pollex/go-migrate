package migration

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

func (m *Migrator) assertMigrationTable() error {
	_, err := m.pg.Exec(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS _meta_migrations(
			id 							SERIAL 			PRIMARY KEY,
			applied_at 			timestamp 	NOT NULL DEFAULT(NOW()),
			current_ix			int					NOT NULL,
			ix 							int 				NOT NULL,
			dir 						int 				NOT NULL,
			label 					varchar(250)
		)`,
	)
	if err != nil {
		return fmt.Errorf("Could not check/create migrations table in database:\n%s", err)
	}

	return nil
}

//
func (m *Migrator) apply(p *Pair, dir migrationDirection) error {
	var err error
	var sql string
	var mig *Single
	var nextIX int

	if dir == DirUp {
		mig = p.up
		nextIX = m.ix + 1
		fmt.Printf("[APPLY] migration %d %s...", mig.ix, mig.label)
	} else {
		mig = p.down
		nextIX = m.ix - 1
		fmt.Printf("[UNDO] migration %d %s...", mig.ix, mig.label)
	}

	sql, err = mig.readSQL()
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := m.pg.Begin(context.Background())
	if err != nil {
		return err
	}

	// Apply migrations
	_, err = tx.Exec(context.Background(), sql)
	if err != nil {
		return err
	}

	// Insert migration info
	_, err = tx.Exec(
		context.Background(),
		`INSERT INTO _meta_migrations(current_ix, ix, dir, label) VALUES ($1, $2, $3, $4)`,
		nextIX,
		mig.ix,
		mig.dir,
		mig.label,
	)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	// End transaction
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	// Update current ix
	m.ix = nextIX

	fmt.Print("done\n")

	return nil
}

// MigrateTo migrations to given connection
func (m *Migrator) MigrateTo(target int) error {
	var dir migrationDirection
	if (target - m.ix) < 0 {
		dir = DirDown
	} else {
		dir = DirUp
	}

	// Sanity checks
	if target > len(m.migrations) {
		fmt.Printf("?? Trying to migrate to %d, but highest available is %d\n", target, len(m.migrations))
		target = len(m.migrations)
	} else if target < 0 {
		fmt.Printf("?? Trying to migrate below 0, defaulting to 0\n")
		target = 0
	}
	fmt.Printf("Migrating from %d to %d\n", m.ix, target)

	// Apply migrations until destination is achieved
	for m.ix != target {
		mig := m.migrations[m.ix+int(dir)]
		err := m.apply(mig, dir)
		if err != nil {
			return err
		}
	}

	return nil
}

// MigrateAll will migrate the database to the most up to date migration
func (m *Migrator) MigrateAll() error {
	target := len(m.migrations)
	return m.MigrateTo(target)
}

// MigrateRelative will migrate a set amount forward or backwards
// if `n` is positive, it wil migrate forward
// if `n` is negative, it will migrate back
func (m *Migrator) MigrateRelative(n int) error {
	target := m.ix + n
	return m.MigrateTo(target)
}

// IX the current applied migration ID
func (m *Migrator) IX() int {
	return m.ix
}

// NewMigrator creates a new migrator
func NewMigrator(pg *pgx.Conn, path string) (*Migrator, error) {
	m := &Migrator{
		pg:   pg,
		path: path,
	}

	// Ensure table exists
	err := m.assertMigrationTable()
	if err != nil {
		return nil, err
	}

	// Find latest migration
	var ix int
	err = pg.QueryRow(context.Background(), `SELECT current_ix FROM _meta_migrations ORDER BY id DESC LIMIT 1`).Scan(&ix)
	if err == pgx.ErrNoRows {
		ix = 0
	} else if err != nil {
		return nil, err
	}
	m.ix = ix

	// Find migration files
	files := findInDir(m.path)
	migrations, err := pair(files)
	if err != nil {
		return nil, err
	}
	m.migrations = migrations

	// Do we have atleast the same amount of migrations available?
	if m.ix > len(m.migrations) {
		return nil, fmt.Errorf("Database is currently at %d, but our migrations only go up to %d", m.ix, len(m.migrations))
	}

	return m, nil
}
