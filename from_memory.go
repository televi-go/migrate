package migrate

import "database/sql"

type InMemory struct {
	Name       string
	Statements string
}

func RunInMemory(db *sql.DB, migrations []InMemory, dialect dialect) error {
	var migrationObjects []migration
	for _, memory := range migrations {
		migrationObjects = append(migrationObjects, migration{qualifiedName: memory.Name, statements: getStatements(memory.Statements)})
	}
	return runMigrationsCore(db, migrationObjects, dialect)
}
