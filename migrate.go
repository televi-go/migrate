package migrate

import (
	"database/sql"
	"golang.org/x/exp/slices"
)

type dialect int

const (
	MysqlDialect dialect = iota
)

var checkMigrationExistenceMap = map[dialect]string{
	MysqlDialect: "select exists (select name from migrations where name = ?)",
}

var insertMigrationMap = map[dialect]string{
	MysqlDialect: "insert into migrations values (?)",
}

func createMigrationsTable(db *sql.DB) error {
	statement := `
create table if not exists migrations(
    name varchar(64) not null primary key
)
`
	_, err := db.Exec(statement)
	return err
}

func runMigration(db *sql.DB, migration migration, dialect dialect) (err error) {
	row := db.QueryRow(checkMigrationExistenceMap[dialect], migration.qualifiedName)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer func(tx *sql.Tx) {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		_, err = tx.Exec(insertMigrationMap[dialect], migration.qualifiedName)
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}(tx)

	for _, statement := range migration.statements {
		_, err = tx.Exec(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

func runMigrationsCore(db *sql.DB, migrations []migration, dialect dialect) error {
	err := createMigrationsTable(db)
	if err != nil {
		return err
	}
	slices.SortFunc(migrations, func(a, b migration) bool {
		return a.Name() < b.Name()
	})
	for _, migration := range migrations {
		err = runMigration(db, migration, dialect)
		if err != nil {
			return err
		}
	}
	return nil
}

func RunMigrationsUp(db *sql.DB, path string, dialect dialect) error {
	migrations, err := traverseDir(path, "")
	if err != nil {
		return err
	}
	return runMigrationsCore(db, migrations, dialect)
}
