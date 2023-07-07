package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

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

func processEntryEmbed(fs embed.FS, path string, entry fs.DirEntry, db *sql.DB, dialect dialect) (result []migration, err error) {
	subDirPath := fmt.Sprintf("%s%c%s", path, os.PathSeparator, entry.Name())
	if entry.IsDir() {
		var entries []os.DirEntry
		entries, err = fs.ReadDir(subDirPath)
		if err != nil {
			return nil, err
		}
		for _, dirEntry := range entries {
			res, err := processEntryEmbed(fs, subDirPath, dirEntry, db, dialect)
			if err != nil {
				return nil, err
			}
			result = append(result, res...)
		}
		return
	}

	if !strings.HasSuffix(subDirPath, ".up.sql") {
		return nil, nil
	}

	content, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result = append(result, migration{qualifiedName: subDirPath, statements: getStatements(string(content))})
	return
}

func RunInEmbed(db *sql.DB, rootName string, fs embed.FS, dialect dialect) error {
	entries, err := fs.ReadDir(rootName)
	if err != nil {
		return err
	}
	var migrations []migration
	for _, entry := range entries {
		entryMigrations, err := processEntryEmbed(fs,
			fmt.Sprintf("%s/%s", rootName, entry.Name()),
			entry,
			db,
			dialect,
		)
		if err != nil {
			return err
		}
		migrations = append(migrations, entryMigrations...)
	}

	return runMigrationsCore(db, migrations, dialect)
}
