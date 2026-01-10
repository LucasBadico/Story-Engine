package relationmaps

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed assets/*.json
var assets embed.FS

func TypesJSON() ([]byte, error) {
	return assets.ReadFile("assets/relation.types.json")
}

func MapJSON(entityType string) ([]byte, error) {
	if entityType == "" {
		return nil, fmt.Errorf("entity type is required")
	}
	path := fmt.Sprintf("assets/%s.relation.map.json", entityType)
	return assets.ReadFile(path)
}

func ListMapFiles() ([]string, error) {
	entries, err := fs.ReadDir(assets, "assets")
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "relation.types.json" {
			continue
		}
		files = append(files, name)
	}
	return files, nil
}
