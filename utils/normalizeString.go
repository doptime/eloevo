package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/doptime/eloevo/config"
	"github.com/samber/lo"
)

func NormalizeFilename(name string) string {
	for found := true; found; {
		found = false
		for _, r := range []string{"Pathname", "pathname", "Path", "./", "src/", "app/", "a/", "b/"} {
			if strings.HasPrefix(name, r) {
				name = strings.TrimPrefix(name, r)
				found = true
			}
		}
	}
	for _, item := range config.EvoRealms {
		if strings.HasPrefix(name, item.Name) {
			name = "/" + item.Name + strings.TrimPrefix(name, item.Name)
		}
	}
	return name
}

func ToLocalEvoFile(path string) (string, *config.EvoRealm) {
	path = NormalizeFilename(path)
	realm, found := lo.Find(config.EvoRealms, func(r *config.EvoRealm) bool {
		return strings.HasPrefix(path, "/"+r.Name) && r.Enable
	})
	return lo.Ternary(found && realm != nil, filepath.Join(realm.RootPath, strings.TrimPrefix(path, "/"+realm.Name)), path), realm
}

func NormalizeFileContent(s string) string {
	s = strings.Replace(s, "use client;\n", "'use client';", 1)
	return s
}
func LineNumberedFileContent(s string) string {
	s = NormalizeFileContent(s)
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = fmt.Sprintf("%d: %s", i+1, lines[i])
	}
	return strings.Join(lines, "\n")
}
