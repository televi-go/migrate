package migrate

import (
	"database/sql"
	"embed"
	"slices"
	"strings"
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
	slices.SortFunc(migrations, func(a, b migration) int {
		if a.Name() == b.Name() {
			return 0
		}
		if a.Name() < b.Name() {
			return -1
		}
		return 1
	})
	for _, migration := range migrations {
		err = runMigration(db, migration, dialect)
		if err != nil {
			return err
		}
	}
	return nil
}

type MigrationExecutor interface {
	IsMigrationAdded(name string) (bool, error)
	GetManager() (TransactionManager, error)
}

type TransactionManager interface {
	Execute(statement string) error
	AddMigration(name string) error
	Commit() error
	Rollback()
}

type migrationWrap struct {
	MigrationExecutor
}

func (m migrationWrap) checkAndExecute(content, name string) (err error) {

	isAdded, err := m.IsMigrationAdded(name)
	if err != nil {
		return err
	}
	if isAdded {
		return nil
	}
	manager, err := m.GetManager()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			manager.Rollback()
			return
		}
		err = manager.Commit()
	}()
	statements := strings.Split(content, ";")
	for _, statement := range statements {
		if isEmptyStatement(statement) {
			continue
		}
		err := manager.Execute(statement)
		if err != nil {
			return err
		}
	}

	err = manager.AddMigration(name)
	return err
}

func RunMigrationsWithExecutor(fs embed.FS, executor MigrationExecutor, rootDir string) error {
	return traverseFS(fs, rootDir, migrationWrap{executor}.checkAndExecute)
}

func RunMigrationsUp(db *sql.DB, path string, dialect dialect) error {
	migrations, err := traverseDir(path, "")
	if err != nil {
		return err
	}
	return runMigrationsCore(db, migrations, dialect)
}
