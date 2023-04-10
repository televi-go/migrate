package migrate

import (
	"fmt"
	"os"
	"strings"
)

type migration struct {
	qualifiedName string
	statements    []string
}

func getStatements(content string) []string {
	raw := strings.Split(content, ";")
	var result []string
	for _, statement := range raw {
		result = append(result, statement)
	}
	return result
}

func traverseDir(path string, prefix string) ([]migration, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("migrate: unable to read folder %s:%w", path, err)
	}
	var result []migration
	for _, entry := range dirEntries {
		if entry.IsDir() {
			subDirResult, err := traverseDir(
				fmt.Sprintf("%s%c%s", path, os.PathSeparator, entry.Name()),
				prefix+entry.Name()+".",
			)
			if err != nil {
				return nil, err
			}
			result = append(result, subDirResult...)
		}

		if !strings.HasSuffix(entry.Name(), ".up.sql") {
			fmt.Printf("migrate: skipping file %s\n", entry.Name())
			continue
		}
		migrationName := strings.TrimSuffix(entry.Name(), ".up.sql")
		qualifiedName := fmt.Sprintf("%s%s", prefix, migrationName)
		contents, err := os.ReadFile(fmt.Sprintf("%s%c%s", path, os.PathSeparator, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("migrate: cannot read file %s: %w", entry.Name(), err)
		}
		result = append(result, migration{
			qualifiedName: qualifiedName,
			statements:    getStatements(string(contents)),
		})
	}
	return result, nil
}
