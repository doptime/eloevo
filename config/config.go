package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	dconfig "github.com/doptime/config"
)

type EvoRealm struct {
	Name      string
	Enable    bool
	RootPath  string
	SkipFiles string
	SkipDirs  string
}

var EvoRealms []*EvoRealm

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

// LoadFilesToMemory loads all JSON configuration files from the specified directory into memory.
func (e *EvoRealm) LoadRealmFiles() (files []*FileData, err error) {
	// Check if the directory exists
	info, err := os.Stat(e.RootPath)
	if os.IsNotExist(err) {
		log.Printf("Directory does not exist: %s", e.RootPath)
		return files, err
	}
	if !info.IsDir() {
		log.Printf("Provided path is not a directory: %s", e.RootPath)
		return files, err
	}
	if !e.Enable {
		log.Printf("EvoRealm disabled: %s", e.Name)
		return files, nil
	}

	skipdirs := strings.Split(e.SkipDirs, ",")
	skipfiles := strings.Split(e.SkipFiles, ",")

	// Walk through the directory
	err = filepath.Walk(e.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			// Skip directories
			for _, skipdir := range skipdirs {
				if len(skipdir) > 0 && strings.Contains(path, skipdir) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Skip files
		for _, skipfile := range skipfiles {
			if len(skipfile) > 0 && strings.Contains(path, skipfile) {
				return nil
			}
		}

		// Read the file content
		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Error reading file %q: %v\n", path, err)
			return err
		}
		files = append(files, &FileData{Path: path, Realm: e, Content: string(content)})
		return nil
	})

	if err != nil {
		log.Printf("Error walking through directory %s: %v", e.RootPath, err)
		return files, err
	}

	log.Printf("All configuration files loaded successfully from directory: %s", e.RootPath)
	return files, nil
}
func LoadRealmsFiles() (files []*FileData, err error) {
	for _, evoRealm := range EvoRealms {
		if evoRealm.Enable {
			_files, _ := evoRealm.LoadRealmFiles()
			files = append(files, _files...)
		}
	}
	return files, nil
}

func init() {
	dconfig.LoadItemFromToml("EvoRealms", &EvoRealms)
}
