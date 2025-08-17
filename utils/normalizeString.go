package utils

import (
	"strings"

	"github.com/doptime/eloevo/config"
)

func NormalizeFilename(name string) string {
	for found := true; found; {
		found = false
		for _, r := range []string{"Pathname", "pathname", "Path", "./", "src/", "app/"} {
			if strings.HasPrefix(name, r) {
				name = strings.TrimPrefix(name, r)
				found = true
			}
		}
	}
	for _, item := range config.EvoRealms {
		if strings.HasPrefix(name, item.Name) {
			name = "$" + item.Name + strings.TrimPrefix(name, item.Name)
		}
	}
	return name
}
func NormalizeFileContent(s string) string {
	s = strings.Replace(s, "use client;\n", "'use client';", 1)
	return s
}
