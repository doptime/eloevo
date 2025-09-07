package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	dconfig "github.com/doptime/config"
	gitignore "github.com/sabhiram/go-gitignore"
)

type EvoRealm struct {
	Name      string
	Enable    bool
	RootPath  string
	GitIgnore string
}

var EvoRealms map[string]*EvoRealm

type FileData struct {
	Path    string
	Realm   *EvoRealm
	Content string
}

func (f *FileData) RealmName() string {
	return strings.Replace(f.Path, f.Realm.RootPath, f.Realm.Name, -1)
}

func (f *FileData) String() string {
	return "\n\nPath: " + f.RealmName() + "\nContent: \n" + f.Content + "\nEOF\n"
}
func DefaultRealmPath() string {
	for _, realm := range EvoRealms {
		if len(realm.Name) > 0 && realm.Enable {
			return realm.RootPath
		}
	}
	fmt.Println("No default realm found in config")
	return ""
}
func (evoRealm *EvoRealm) WalkDir(fn func(path, relativePath string, info os.FileInfo) error) {
	// Check if the directory exists
	info, err := os.Stat(evoRealm.RootPath)
	if os.IsNotExist(err) {
		log.Printf("Directory does not exist: %s", evoRealm.RootPath)
		return
	}
	if !info.IsDir() {
		log.Printf("Provided path is not a directory: %s", evoRealm.RootPath)
		return
	}
	ignoreLines := strings.Split(evoRealm.GitIgnore, ",")
	ignorer := gitignore.CompileIgnoreLines(ignoreLines...)

	// Walk through the directory
	filepath.Walk(evoRealm.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}

		if ignorer.MatchesPath(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}

		relativePath := "/" + evoRealm.Name + strings.TrimPrefix(path, evoRealm.RootPath)
		fn(path, relativePath, info)

		return nil
	})
}

func init() {
	var EvoRealmsArray []*EvoRealm
	dconfig.LoadItemFromToml("EvoRealms", &EvoRealmsArray)
	EvoRealms = make(map[string]*EvoRealm)
	for _, realm := range EvoRealmsArray {
		EvoRealms[realm.Name] = realm
	}
}
