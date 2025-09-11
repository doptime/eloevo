package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	dconfig "github.com/doptime/config"
	"github.com/dustin/go-humanize"
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/samber/lo"
)

type EvoRealm struct {
	Name      string
	Enable    bool
	RootPath  string
	GitIgnore string
}

func (realm *EvoRealm) RelativePath(path string) string {
	return "/" + realm.Name + strings.TrimPrefix(path, realm.RootPath)
}
func (realm *EvoRealm) GitDiffFile(relativePath string) string {
	filename := strings.TrimLeft(relativePath, "/"+realm.Name)
	filename = strings.Join(strings.Split(filename, "/")[1:], "--")

	dir := strings.TrimRight(realm.RootPath, "/") + "/.evo/"
	os.MkdirAll(dir, 0755)
	gitdiffHistoryFile := dir + "/" + filename
	return gitdiffHistoryFile
}

func (realm *EvoRealm) LoadOneXmlFile(path, gitdiffFile string) string {
	relativePath := realm.RelativePath(path)
	// Read the file content
	contentWithXmlTag := ""
	bytes, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	content := strings.TrimSpace(string(bytes))
	LineNumberedFileContent := func(s string, LineNumberStart ...int) string {
		s = strings.Replace(s, "use client;\n", "'use client';", 1)
		lines := strings.Split(s, "\n")
		sequenceBase := lo.Ternary(len(LineNumberStart) > 0, LineNumberStart[0], 0)
		for i := range lines {
			lines[i] = fmt.Sprintf("%d: %s", i+sequenceBase, lines[i])
		}
		return strings.Join(lines, "\n")
	}

	fileSz := "\n<file-size>" + humanize.Bytes(uint64(len(content))) + "</file-size>"
	if binaryFile := strings.Contains(string(content), "\x00") || len(content) == 0; binaryFile {
		contentWithXmlTag = "<file-content>{BinaryFileData}</file-content>"
	} else {
		contentWithXmlTag = "\n<file-content>\n" + LineNumberedFileContent(string(content), 1) + "\n</file-content>"

	}
	commitStr := ""

	type GitDiffs struct {
		Diffs []string `toml:"Diffs"`
	}
	if gitdiffFile != "" {
		diffs := GitDiffs{}
		toml.DecodeFile(path, &diffs)
		if len(diffs.Diffs) > 0 {
			var gitdiffs strings.Builder
			for i := 0; i < len(diffs.Diffs) && i < 5; i++ {
				gitdiffs.WriteString(diffs.Diffs[i] + "\n")
			}
			commitStr += "\n<git-commits-unified-diff-file>\n" + gitdiffs.String() + "\n</git-commits-unified-diff-file>"
		}
	}

	fileinfo := fmt.Sprint("\n<file>\n<file-name>", relativePath, "</file-name>"+fileSz+commitStr, contentWithXmlTag, "\n</file>\n")
	return fileinfo
}

// 将LoadAllEvoProjects 放到utils ,部分出于
func (realm *EvoRealm) LoadProjectFiles(fileKeepMap map[string]bool) string {
	var allFileInfo strings.Builder

	realm.WalkDir(func(path, relativePath string, info os.FileInfo) (e error) {
		fmt.Printf("Processing file: %s\n", path)
		if len(fileKeepMap) > 0 {
			if _, ok := fileKeepMap[relativePath]; !ok {
				return nil
			}
		}
		// gitdiffs file, locates in <realm.RootPath>/.evo
		//ensure .evo dir exists
		dir := strings.TrimRight(realm.RootPath, "/") + "/.evo"
		os.MkdirAll(dir, 0755)
		gitdiffHistoryFile := dir + "/" + strings.Join(strings.Split(relativePath, "/")[1:], "--")

		allFileInfo.WriteString(realm.LoadOneXmlFile(path, gitdiffHistoryFile))
		return nil
	})
	return allFileInfo.String()
}

// 将LoadAllEvoProjects 放到utils ,部分出于
func LoadAllEvoProjects(KeepFileNames ...[]string) string {
	var allFileInfo strings.Builder
	fileKeepMap := map[string]bool{}

	for _, fn := range KeepFileNames {
		for _, f := range fn {
			fileKeepMap[f] = true
		}
	}

	for _, realm := range lo.Filter(lo.Values(EvoRealms), func(realm *EvoRealm, _ int) bool { return realm.Enable }) {
		allFileInfo.WriteString(realm.LoadProjectFiles(fileKeepMap))
	}
	return allFileInfo.String()
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
