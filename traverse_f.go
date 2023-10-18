package migrate

import (
	"embed"
	"strings"
)

func traverseFS(fs embed.FS, onFile func(content, name string) error) error {
	files, err := fs.ReadDir(".")
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		content, err := fs.ReadFile(file.Name())
		if err != nil {
			return err
		}
		err = onFile(string(content), strings.TrimSuffix(file.Name(), ".up.sql"))
		if err != nil {
			return err
		}
	}
	return nil
}
